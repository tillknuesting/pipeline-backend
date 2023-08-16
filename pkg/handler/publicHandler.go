package handler

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gofrs/uuid"
	"github.com/gogo/status"
	"github.com/iancoleman/strcase"
	"go.einride.tech/aip/filtering"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"

	fieldmask_utils "github.com/mennanov/fieldmask-utils"

	"github.com/instill-ai/connector-backend/pkg/repository"
	"github.com/instill-ai/pipeline-backend/internal/resource"
	"github.com/instill-ai/pipeline-backend/pkg/datamodel"
	"github.com/instill-ai/pipeline-backend/pkg/logger"
	"github.com/instill-ai/pipeline-backend/pkg/operator"
	"github.com/instill-ai/pipeline-backend/pkg/service"
	"github.com/instill-ai/pipeline-backend/pkg/utils"
	"github.com/instill-ai/x/checkfield"
	"github.com/instill-ai/x/paginate"
	"github.com/instill-ai/x/sterr"

	custom_otel "github.com/instill-ai/pipeline-backend/pkg/logger/otel"
	mgmtPB "github.com/instill-ai/protogen-go/base/mgmt/v1alpha"
	healthcheckPB "github.com/instill-ai/protogen-go/common/healthcheck/v1alpha"
	pipelinePB "github.com/instill-ai/protogen-go/vdp/pipeline/v1alpha"
)

var tracer = otel.Tracer("pipeline-backend.public-handler.tracer")

// PublicHandler handles public API
type PublicHandler struct {
	pipelinePB.UnimplementedPipelinePublicServiceServer
	service  service.Service
	operator operator.Operator
}

type Streamer interface {
	Context() context.Context
}

type TriggerPipelineRequestInterface interface {
	GetName() string
}

// NewPublicHandler initiates a handler instance
func NewPublicHandler(ctx context.Context, s service.Service) pipelinePB.PipelinePublicServiceServer {
	datamodel.InitJSONSchema(ctx)
	return &PublicHandler{
		service:  s,
		operator: operator.InitOperator(),
	}
}

// GetService returns the service
func (h *PublicHandler) GetService() service.Service {
	return h.service
}

// SetService sets the service
func (h *PublicHandler) SetService(s service.Service) {
	h.service = s
}

func (h *PublicHandler) Liveness(ctx context.Context, req *pipelinePB.LivenessRequest) (*pipelinePB.LivenessResponse, error) {
	return &pipelinePB.LivenessResponse{
		HealthCheckResponse: &healthcheckPB.HealthCheckResponse{
			Status: healthcheckPB.HealthCheckResponse_SERVING_STATUS_SERVING,
		},
	}, nil
}

func (h *PublicHandler) Readiness(ctx context.Context, req *pipelinePB.ReadinessRequest) (*pipelinePB.ReadinessResponse, error) {
	return &pipelinePB.ReadinessResponse{
		HealthCheckResponse: &healthcheckPB.HealthCheckResponse{
			Status: healthcheckPB.HealthCheckResponse_SERVING_STATUS_SERVING,
		},
	}, nil
}

func (h *PublicHandler) ListOperatorDefinitions(ctx context.Context, req *pipelinePB.ListOperatorDefinitionsRequest) (resp *pipelinePB.ListOperatorDefinitionsResponse, err error) {
	ctx, span := tracer.Start(ctx, "ListOperatorDefinitions",
		trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	logger, _ := logger.GetZapLogger(ctx)

	resp = &pipelinePB.ListOperatorDefinitionsResponse{}
	pageSize := req.GetPageSize()
	pageToken := req.GetPageToken()
	isBasicView := (req.GetView() == pipelinePB.View_VIEW_BASIC) || (req.GetView() == pipelinePB.View_VIEW_UNSPECIFIED)

	prevLastUid := ""

	if pageToken != "" {
		_, prevLastUid, err = paginate.DecodeToken(pageToken)
		if err != nil {
			st, err := sterr.CreateErrorBadRequest(
				fmt.Sprintf("[db] list operator error: %s", err.Error()),
				[]*errdetails.BadRequest_FieldViolation{
					{
						Field:       "page_token",
						Description: fmt.Sprintf("Invalid page token: %s", err.Error()),
					},
				},
			)
			if err != nil {
				logger.Error(err.Error())
			}
			return nil, st.Err()
		}
	}

	if pageSize == 0 {
		pageSize = repository.DefaultPageSize
	} else if pageSize > repository.MaxPageSize {
		pageSize = repository.MaxPageSize
	}

	defs := h.operator.ListOperatorDefinitions()

	startIdx := 0
	lastUid := ""
	for idx, def := range defs {
		if def.Uid == prevLastUid {
			startIdx = idx + 1
			break
		}
	}

	page := []*pipelinePB.OperatorDefinition{}
	for i := 0; i < int(pageSize) && startIdx+i < len(defs); i++ {
		def := proto.Clone(defs[startIdx+i]).(*pipelinePB.OperatorDefinition)
		page = append(page, def)
		lastUid = def.Uid
	}

	nextPageToken := ""

	if startIdx+len(page) < len(defs) {
		nextPageToken = paginate.EncodeToken(time.Time{}, lastUid)
	}
	for _, def := range page {
		def.Name = fmt.Sprintf("operator-definitions/%s", def.Id)
		if isBasicView {
			def.Spec = nil
		}
		resp.OperatorDefinitions = append(
			resp.OperatorDefinitions,
			def)
	}
	resp.NextPageToken = nextPageToken
	resp.TotalSize = int64(len(defs))

	logger.Info("ListOperatorDefinitions")

	return resp, nil
}

func (h *PublicHandler) GetOperatorDefinition(ctx context.Context, req *pipelinePB.GetOperatorDefinitionRequest) (resp *pipelinePB.GetOperatorDefinitionResponse, err error) {
	ctx, span := tracer.Start(ctx, "GetOperatorDefinition",
		trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	logger, _ := logger.GetZapLogger(ctx)

	resp = &pipelinePB.GetOperatorDefinitionResponse{}

	var connID string

	if connID, err = resource.GetRscNameID(req.GetName()); err != nil {
		span.SetStatus(1, err.Error())
		return resp, err
	}
	isBasicView := (req.GetView() == pipelinePB.View_VIEW_BASIC) || (req.GetView() == pipelinePB.View_VIEW_UNSPECIFIED)

	dbDef, err := h.operator.GetOperatorDefinitionById(connID)
	if err != nil {
		span.SetStatus(1, err.Error())
		return resp, err
	}
	resp.OperatorDefinition = proto.Clone(dbDef).(*pipelinePB.OperatorDefinition)
	if isBasicView {
		resp.OperatorDefinition.Spec = nil
	}
	resp.OperatorDefinition.Name = fmt.Sprintf("operator-definitions/%s", resp.OperatorDefinition.GetId())

	logger.Info("GetOperatorDefinition")
	return resp, nil
}

func (h *PublicHandler) CreatePipeline(ctx context.Context, req *pipelinePB.CreatePipelineRequest) (*pipelinePB.CreatePipelineResponse, error) {

	eventName := "CreatePipeline"

	ctx, span := tracer.Start(ctx, eventName,
		trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	logUUID, _ := uuid.NewV4()

	logger, _ := logger.GetZapLogger(ctx)

	// Validate JSON Schema
	// if err := datamodel.ValidatePipelineJSONSchema(req.GetPipeline()); err != nil {
	// 	span.SetStatus(1, err.Error())
	// 	return nil, status.Error(codes.InvalidArgument, err.Error())
	// }

	// Return error if REQUIRED fields are not provided in the requested payload pipeline resource
	if err := checkfield.CheckRequiredFields(req.Pipeline, append(createRequiredFields, immutableFields...)); err != nil {
		span.SetStatus(1, err.Error())
		return &pipelinePB.CreatePipelineResponse{}, status.Error(codes.InvalidArgument, err.Error())
	}

	// Set all OUTPUT_ONLY fields to zero value on the requested payload pipeline resource
	if err := checkfield.CheckCreateOutputOnlyFields(req.Pipeline, outputOnlyFields); err != nil {
		span.SetStatus(1, err.Error())
		return &pipelinePB.CreatePipelineResponse{}, status.Error(codes.InvalidArgument, err.Error())
	}

	// Return error if resource ID does not follow RFC-1034
	if err := checkfield.CheckResourceID(req.Pipeline.GetId()); err != nil {
		span.SetStatus(1, err.Error())
		return &pipelinePB.CreatePipelineResponse{}, status.Error(codes.InvalidArgument, err.Error())
	}

	owner, err := resource.GetOwner(ctx, h.service.GetMgmtPrivateServiceClient(), h.service.GetRedisClient())
	if err != nil {
		span.SetStatus(1, err.Error())
		return &pipelinePB.CreatePipelineResponse{}, err
	}

	dbPipeline, err := h.service.CreatePipeline(owner, PBToDBPipeline(ctx, owner.GetName(), req.GetPipeline()))
	if err != nil {
		span.SetStatus(1, err.Error())
		// Manually set the custom header to have a StatusBadRequest http response for REST endpoint
		if err := grpc.SetHeader(ctx, metadata.Pairs("x-http-code", strconv.Itoa(http.StatusBadRequest))); err != nil {
			return &pipelinePB.CreatePipelineResponse{Pipeline: &pipelinePB.Pipeline{Recipe: &pipelinePB.Recipe{}}}, err
		}
		return &pipelinePB.CreatePipelineResponse{Pipeline: &pipelinePB.Pipeline{}}, err
	}

	pbPipeline := DBToPBPipeline(ctx, dbPipeline)
	resp := pipelinePB.CreatePipelineResponse{
		Pipeline: pbPipeline,
	}

	// Manually set the custom header to have a StatusCreated http response for REST endpoint
	if err := grpc.SetHeader(ctx, metadata.Pairs("x-http-code", strconv.Itoa(http.StatusCreated))); err != nil {
		span.SetStatus(1, err.Error())
		return nil, err
	}

	logger.Info(string(custom_otel.NewLogMessage(
		span,
		logUUID.String(),
		owner,
		eventName,
		custom_otel.SetEventResource(dbPipeline),
	)))

	return &resp, nil
}

func (h *PublicHandler) ListPipelines(ctx context.Context, req *pipelinePB.ListPipelinesRequest) (*pipelinePB.ListPipelinesResponse, error) {

	eventName := "ListPipelines"

	ctx, span := tracer.Start(ctx, eventName,
		trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	logUUID, _ := uuid.NewV4()

	logger, _ := logger.GetZapLogger(ctx)

	isBasicView := (req.GetView() == pipelinePB.View_VIEW_BASIC) || (req.GetView() == pipelinePB.View_VIEW_UNSPECIFIED)

	owner, err := resource.GetOwner(ctx, h.service.GetMgmtPrivateServiceClient(), h.service.GetRedisClient())
	if err != nil {
		span.SetStatus(1, err.Error())
		return &pipelinePB.ListPipelinesResponse{}, err
	}

	var state pipelinePB.Pipeline_State
	declarations, err := filtering.NewDeclarations([]filtering.DeclarationOption{
		filtering.DeclareStandardFunctions(),
		filtering.DeclareFunction("time.now", filtering.NewFunctionOverload("time.now", filtering.TypeTimestamp)),
		filtering.DeclareIdent("uid", filtering.TypeString),
		filtering.DeclareIdent("id", filtering.TypeString),
		filtering.DeclareIdent("description", filtering.TypeString),
		// only support "recipe.components.resource_name" for now
		filtering.DeclareIdent("recipe", filtering.TypeMap(filtering.TypeString, filtering.TypeMap(filtering.TypeString, filtering.TypeString))),
		filtering.DeclareEnumIdent("state", state.Type()),
		filtering.DeclareIdent("owner", filtering.TypeString),
		filtering.DeclareIdent("create_time", filtering.TypeTimestamp),
		filtering.DeclareIdent("update_time", filtering.TypeTimestamp),
	}...)
	if err != nil {
		span.SetStatus(1, err.Error())
		return &pipelinePB.ListPipelinesResponse{}, err
	}

	filter, err := filtering.ParseFilter(req, declarations)
	if err != nil {
		span.SetStatus(1, err.Error())
		return &pipelinePB.ListPipelinesResponse{}, err
	}

	dbPipelines, totalSize, nextPageToken, err := h.service.ListPipelines(owner, req.GetPageSize(), req.GetPageToken(), isBasicView, filter)
	if err != nil {
		span.SetStatus(1, err.Error())
		return &pipelinePB.ListPipelinesResponse{}, err
	}

	pbPipelines := []*pipelinePB.Pipeline{}
	for idx := range dbPipelines {
		pbPipelines = append(pbPipelines, DBToPBPipeline(ctx, &dbPipelines[idx]))
	}

	logger.Info(string(custom_otel.NewLogMessage(
		span,
		logUUID.String(),
		owner,
		eventName,
	)))

	resp := pipelinePB.ListPipelinesResponse{
		Pipelines:     pbPipelines,
		NextPageToken: nextPageToken,
		TotalSize:     totalSize,
	}

	return &resp, nil
}

func (h *PublicHandler) GetPipeline(ctx context.Context, req *pipelinePB.GetPipelineRequest) (*pipelinePB.GetPipelineResponse, error) {

	eventName := "GetPipeline"

	ctx, span := tracer.Start(ctx, eventName,
		trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	logUUID, _ := uuid.NewV4()

	logger, _ := logger.GetZapLogger(ctx)

	isBasicView := (req.GetView() == pipelinePB.View_VIEW_BASIC) || (req.GetView() == pipelinePB.View_VIEW_UNSPECIFIED)

	owner, err := resource.GetOwner(ctx, h.service.GetMgmtPrivateServiceClient(), h.service.GetRedisClient())
	if err != nil {
		span.SetStatus(1, err.Error())
		return &pipelinePB.GetPipelineResponse{}, err
	}

	id, err := resource.GetRscNameID(req.GetName())
	if err != nil {
		span.SetStatus(1, err.Error())
		return &pipelinePB.GetPipelineResponse{}, err
	}

	dbPipeline, err := h.service.GetPipelineByID(id, owner, isBasicView)
	if err != nil {
		span.SetStatus(1, err.Error())
		return &pipelinePB.GetPipelineResponse{}, err
	}

	pbPipeline := DBToPBPipeline(ctx, dbPipeline)
	resp := pipelinePB.GetPipelineResponse{
		Pipeline: pbPipeline,
	}

	logger.Info(string(custom_otel.NewLogMessage(
		span,
		logUUID.String(),
		owner,
		eventName,
		custom_otel.SetEventResource(dbPipeline),
	)))

	return &resp, nil
}

func (h *PublicHandler) UpdatePipeline(ctx context.Context, req *pipelinePB.UpdatePipelineRequest) (*pipelinePB.UpdatePipelineResponse, error) {

	eventName := "UpdatePipeline"

	ctx, span := tracer.Start(ctx, eventName,
		trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	logUUID, _ := uuid.NewV4()

	logger, _ := logger.GetZapLogger(ctx)

	owner, err := resource.GetOwner(ctx, h.service.GetMgmtPrivateServiceClient(), h.service.GetRedisClient())
	if err != nil {
		span.SetStatus(1, err.Error())
		return &pipelinePB.UpdatePipelineResponse{}, err
	}

	pbPipelineReq := req.GetPipeline()
	pbUpdateMask := req.GetUpdateMask()

	// Validate the field mask
	if !pbUpdateMask.IsValid(pbPipelineReq) {
		return &pipelinePB.UpdatePipelineResponse{}, status.Error(codes.InvalidArgument, "The update_mask is invalid")
	}

	getResp, err := h.GetPipeline(ctx, &pipelinePB.GetPipelineRequest{Name: pbPipelineReq.GetName()})
	if err != nil {
		span.SetStatus(1, err.Error())
		return &pipelinePB.UpdatePipelineResponse{}, err
	}

	pbUpdateMask, err = checkfield.CheckUpdateOutputOnlyFields(pbUpdateMask, outputOnlyFields)
	if err != nil {
		span.SetStatus(1, err.Error())
		return &pipelinePB.UpdatePipelineResponse{}, status.Error(codes.InvalidArgument, err.Error())
	}

	mask, err := fieldmask_utils.MaskFromProtoFieldMask(pbUpdateMask, strcase.ToCamel)
	if err != nil {
		span.SetStatus(1, err.Error())
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if mask.IsEmpty() {
		return &pipelinePB.UpdatePipelineResponse{
			Pipeline: getResp.GetPipeline(),
		}, nil
	}

	pbPipelineToUpdate := getResp.GetPipeline()
	if pbPipelineToUpdate.State == pipelinePB.Pipeline_STATE_ACTIVE {
		return &pipelinePB.UpdatePipelineResponse{}, status.Error(codes.InvalidArgument, "can not update a active pipeline")
	}

	// Return error if IMMUTABLE fields are intentionally changed
	if err := checkfield.CheckUpdateImmutableFields(pbPipelineReq, pbPipelineToUpdate, immutableFields); err != nil {
		span.SetStatus(1, err.Error())
		return &pipelinePB.UpdatePipelineResponse{}, status.Error(codes.InvalidArgument, err.Error())
	}

	// Only the fields mentioned in the field mask will be copied to `pbPipelineToUpdate`, other fields are left intact
	err = fieldmask_utils.StructToStruct(mask, pbPipelineReq, pbPipelineToUpdate)
	if err != nil {
		span.SetStatus(1, err.Error())
		return &pipelinePB.UpdatePipelineResponse{}, err
	}

	dbPipeline, err := h.service.UpdatePipeline(pbPipelineToUpdate.GetId(), owner, PBToDBPipeline(ctx, owner.GetName(), pbPipelineToUpdate))
	if err != nil {
		span.SetStatus(1, err.Error())
		return &pipelinePB.UpdatePipelineResponse{}, err
	}

	resp := pipelinePB.UpdatePipelineResponse{
		Pipeline: DBToPBPipeline(ctx, dbPipeline),
	}

	logger.Info(string(custom_otel.NewLogMessage(
		span,
		logUUID.String(),
		owner,
		eventName,
		custom_otel.SetEventResource(dbPipeline),
	)))

	return &resp, nil
}

func (h *PublicHandler) DeletePipeline(ctx context.Context, req *pipelinePB.DeletePipelineRequest) (*pipelinePB.DeletePipelineResponse, error) {

	eventName := "DeletePipeline"

	ctx, span := tracer.Start(ctx, eventName,
		trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	logUUID, _ := uuid.NewV4()

	logger, _ := logger.GetZapLogger(ctx)

	owner, err := resource.GetOwner(ctx, h.service.GetMgmtPrivateServiceClient(), h.service.GetRedisClient())
	if err != nil {
		span.SetStatus(1, err.Error())
		return &pipelinePB.DeletePipelineResponse{}, err
	}

	existPipeline, err := h.GetPipeline(ctx, &pipelinePB.GetPipelineRequest{Name: req.GetName()})
	if err != nil {
		span.SetStatus(1, err.Error())
		return &pipelinePB.DeletePipelineResponse{}, err
	}

	if err := h.service.DeletePipeline(existPipeline.GetPipeline().GetId(), owner); err != nil {
		span.SetStatus(1, err.Error())
		return &pipelinePB.DeletePipelineResponse{}, err
	}

	// We need to manually set the custom header to have a StatusCreated http response for REST endpoint
	if err := grpc.SetHeader(ctx, metadata.Pairs("x-http-code", strconv.Itoa(http.StatusNoContent))); err != nil {
		span.SetStatus(1, err.Error())
		return &pipelinePB.DeletePipelineResponse{}, err
	}

	logger.Info(string(custom_otel.NewLogMessage(
		span,
		logUUID.String(),
		owner,
		eventName,
		custom_otel.SetEventResource(existPipeline.GetPipeline()),
	)))

	return &pipelinePB.DeletePipelineResponse{}, nil
}

func (h *PublicHandler) LookUpPipeline(ctx context.Context, req *pipelinePB.LookUpPipelineRequest) (*pipelinePB.LookUpPipelineResponse, error) {

	eventName := "LookUpPipeline"

	ctx, span := tracer.Start(ctx, eventName,
		trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	logUUID, _ := uuid.NewV4()

	logger, _ := logger.GetZapLogger(ctx)

	// Return error if REQUIRED fields are not provided in the requested payload pipeline resource
	if err := checkfield.CheckRequiredFields(req, lookUpRequiredFields); err != nil {
		span.SetStatus(1, err.Error())
		return &pipelinePB.LookUpPipelineResponse{}, status.Error(codes.InvalidArgument, err.Error())
	}

	isBasicView := (req.GetView() == pipelinePB.View_VIEW_BASIC) || (req.GetView() == pipelinePB.View_VIEW_UNSPECIFIED)

	owner, err := resource.GetOwner(ctx, h.service.GetMgmtPrivateServiceClient(), h.service.GetRedisClient())
	if err != nil {
		span.SetStatus(1, err.Error())
		return &pipelinePB.LookUpPipelineResponse{}, err
	}

	uidStr, err := resource.GetPermalinkUID(req.GetPermalink())
	if err != nil {
		span.SetStatus(1, err.Error())
		return &pipelinePB.LookUpPipelineResponse{}, err
	}

	uid, err := uuid.FromString(uidStr)
	if err != nil {
		span.SetStatus(1, err.Error())
		return &pipelinePB.LookUpPipelineResponse{}, err
	}

	dbPipeline, err := h.service.GetPipelineByUID(uid, owner, isBasicView)
	if err != nil {
		span.SetStatus(1, err.Error())
		return &pipelinePB.LookUpPipelineResponse{}, err
	}

	pbPipeline := DBToPBPipeline(ctx, dbPipeline)
	resp := pipelinePB.LookUpPipelineResponse{
		Pipeline: pbPipeline,
	}

	logger.Info(string(custom_otel.NewLogMessage(
		span,
		logUUID.String(),
		owner,
		eventName,
		custom_otel.SetEventResource(dbPipeline),
	)))

	return &resp, nil
}

func (h *PublicHandler) ActivatePipeline(ctx context.Context, req *pipelinePB.ActivatePipelineRequest) (*pipelinePB.ActivatePipelineResponse, error) {

	eventName := "ActivatePipeline"

	ctx, span := tracer.Start(ctx, eventName,
		trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	logUUID, _ := uuid.NewV4()

	logger, _ := logger.GetZapLogger(ctx)

	// Return error if REQUIRED fields are not provided in the requested payload pipeline resource
	if err := checkfield.CheckRequiredFields(req, activateRequiredFields); err != nil {
		span.SetStatus(1, err.Error())
		return &pipelinePB.ActivatePipelineResponse{}, status.Error(codes.InvalidArgument, err.Error())
	}

	owner, err := resource.GetOwner(ctx, h.service.GetMgmtPrivateServiceClient(), h.service.GetRedisClient())
	if err != nil {
		span.SetStatus(1, err.Error())
		return &pipelinePB.ActivatePipelineResponse{}, err
	}

	id, err := resource.GetRscNameID(req.GetName())
	if err != nil {
		span.SetStatus(1, err.Error())
		return &pipelinePB.ActivatePipelineResponse{}, err
	}

	dbPipeline, err := h.service.UpdatePipelineState(id, owner, datamodel.PipelineState(pipelinePB.Pipeline_STATE_ACTIVE))
	if err != nil {
		span.SetStatus(1, err.Error())
		return &pipelinePB.ActivatePipelineResponse{}, err
	}

	resp := pipelinePB.ActivatePipelineResponse{
		Pipeline: DBToPBPipeline(ctx, dbPipeline),
	}

	logger.Info(string(custom_otel.NewLogMessage(
		span,
		logUUID.String(),
		owner,
		eventName,
		custom_otel.SetEventResource(dbPipeline),
	)))

	return &resp, nil
}

func (h *PublicHandler) DeactivatePipeline(ctx context.Context, req *pipelinePB.DeactivatePipelineRequest) (*pipelinePB.DeactivatePipelineResponse, error) {

	eventName := "DeactivatePipeline"

	ctx, span := tracer.Start(ctx, eventName,
		trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	logUUID, _ := uuid.NewV4()

	logger, _ := logger.GetZapLogger(ctx)

	// Return error if REQUIRED fields are not provided in the requested payload pipeline resource
	if err := checkfield.CheckRequiredFields(req, deactivateRequiredFields); err != nil {
		span.SetStatus(1, err.Error())
		return &pipelinePB.DeactivatePipelineResponse{}, status.Error(codes.InvalidArgument, err.Error())
	}

	owner, err := resource.GetOwner(ctx, h.service.GetMgmtPrivateServiceClient(), h.service.GetRedisClient())
	if err != nil {
		span.SetStatus(1, err.Error())
		return &pipelinePB.DeactivatePipelineResponse{}, err
	}

	id, err := resource.GetRscNameID(req.GetName())
	if err != nil {
		span.SetStatus(1, err.Error())
		return &pipelinePB.DeactivatePipelineResponse{}, err
	}

	dbPipeline, err := h.service.UpdatePipelineState(id, owner, datamodel.PipelineState(pipelinePB.Pipeline_STATE_INACTIVE))
	if err != nil {
		span.SetStatus(1, err.Error())
		return &pipelinePB.DeactivatePipelineResponse{}, err
	}

	resp := pipelinePB.DeactivatePipelineResponse{
		Pipeline: DBToPBPipeline(ctx, dbPipeline),
	}

	logger.Info(string(custom_otel.NewLogMessage(
		span,
		logUUID.String(),
		owner,
		eventName,
		custom_otel.SetEventResource(dbPipeline),
	)))

	return &resp, nil
}

func (h *PublicHandler) RenamePipeline(ctx context.Context, req *pipelinePB.RenamePipelineRequest) (*pipelinePB.RenamePipelineResponse, error) {

	eventName := "RenamePipeline"

	ctx, span := tracer.Start(ctx, eventName,
		trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	logUUID, _ := uuid.NewV4()

	logger, _ := logger.GetZapLogger(ctx)

	// Return error if REQUIRED fields are not provided in the requested payload pipeline resource
	if err := checkfield.CheckRequiredFields(req, renameRequiredFields); err != nil {
		span.SetStatus(1, err.Error())
		return &pipelinePB.RenamePipelineResponse{}, status.Error(codes.InvalidArgument, err.Error())
	}

	owner, err := resource.GetOwner(ctx, h.service.GetMgmtPrivateServiceClient(), h.service.GetRedisClient())
	if err != nil {
		span.SetStatus(1, err.Error())
		return &pipelinePB.RenamePipelineResponse{}, err
	}

	id, err := resource.GetRscNameID(req.GetName())
	if err != nil {
		span.SetStatus(1, err.Error())
		return &pipelinePB.RenamePipelineResponse{}, err
	}

	newID := req.GetNewPipelineId()
	if err := checkfield.CheckResourceID(newID); err != nil {
		span.SetStatus(1, err.Error())
		return &pipelinePB.RenamePipelineResponse{}, status.Error(codes.InvalidArgument, err.Error())
	}

	dbPipeline, err := h.service.UpdatePipelineID(id, owner, newID)
	if err != nil {
		span.SetStatus(1, err.Error())
		return &pipelinePB.RenamePipelineResponse{}, err
	}

	resp := pipelinePB.RenamePipelineResponse{
		Pipeline: DBToPBPipeline(ctx, dbPipeline),
	}

	logger.Info(string(custom_otel.NewLogMessage(
		span,
		logUUID.String(),
		owner,
		eventName,
		custom_otel.SetEventResource(dbPipeline),
	)))

	return &resp, nil
}

func (h *PublicHandler) PreTriggerPipeline(ctx context.Context, req TriggerPipelineRequestInterface) (*mgmtPB.User, *datamodel.Pipeline, error) {

	// Return error if REQUIRED fields are not provided in the requested payload pipeline resource
	if err := checkfield.CheckRequiredFields(req, triggerRequiredFields); err != nil {
		return nil, nil, status.Error(codes.InvalidArgument, err.Error())
	}

	owner, err := resource.GetOwner(ctx, h.service.GetMgmtPrivateServiceClient(), h.service.GetRedisClient())
	if err != nil {
		return nil, nil, err
	}

	id, err := resource.GetRscNameID(req.GetName())
	if err != nil {
		return nil, nil, err
	}

	dbPipeline, err := h.service.GetPipelineByID(id, owner, false)
	if err != nil {
		return nil, nil, err
	}

	return owner, dbPipeline, nil

}

func (h *PublicHandler) TriggerPipeline(ctx context.Context, req *pipelinePB.TriggerPipelineRequest) (*pipelinePB.TriggerPipelineResponse, error) {

	startTime := time.Now()
	eventName := "TriggerPipeline"

	ctx, span := tracer.Start(ctx, eventName,
		trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	logUUID, _ := uuid.NewV4()

	logger, _ := logger.GetZapLogger(ctx)

	owner, dbPipeline, err := h.PreTriggerPipeline(ctx, req)
	if err != nil {
		span.SetStatus(1, err.Error())
		return &pipelinePB.TriggerPipelineResponse{}, err
	}

	dataPoint := utils.UsageMetricData{
		OwnerUID:           *owner.Uid,
		TriggerMode:        mgmtPB.Mode_MODE_SYNC,
		PipelineID:         dbPipeline.ID,
		PipelineUID:        dbPipeline.UID.String(),
		PipelineTriggerUID: logUUID.String(),
		TriggerTime:        startTime.Format(time.RFC3339Nano),
	}

	resp, err := h.service.TriggerPipeline(ctx, req, owner, dbPipeline, logUUID.String())
	if err != nil {
		span.SetStatus(1, err.Error())
		dataPoint.ComputeTimeDuration = time.Since(startTime).Seconds()
		dataPoint.Status = mgmtPB.Status_STATUS_ERRORED
		_ = h.service.WriteNewDataPoint(ctx, dataPoint)
		return &pipelinePB.TriggerPipelineResponse{}, err
	}

	logger.Info(string(custom_otel.NewLogMessage(
		span,
		logUUID.String(),
		owner,
		eventName,
		custom_otel.SetEventResource(dbPipeline),
	)))

	dataPoint.ComputeTimeDuration = time.Since(startTime).Seconds()
	dataPoint.Status = mgmtPB.Status_STATUS_COMPLETED
	if err := h.service.WriteNewDataPoint(ctx, dataPoint); err != nil {
		logger.Warn(err.Error())
	}

	return resp, nil
}

func (h *PublicHandler) TriggerAsyncPipeline(ctx context.Context, req *pipelinePB.TriggerAsyncPipelineRequest) (*pipelinePB.TriggerAsyncPipelineResponse, error) {

	eventName := "TriggerAsyncPipeline"

	ctx, span := tracer.Start(ctx, eventName,
		trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	logUUID, _ := uuid.NewV4()

	logger, _ := logger.GetZapLogger(ctx)

	owner, dbPipeline, err := h.PreTriggerPipeline(ctx, req)
	if err != nil {
		span.SetStatus(1, err.Error())
		return &pipelinePB.TriggerAsyncPipelineResponse{}, err
	}

	resp, err := h.service.TriggerAsyncPipeline(ctx, req, logUUID.String(), owner, dbPipeline)
	if err != nil {
		span.SetStatus(1, err.Error())
		return &pipelinePB.TriggerAsyncPipelineResponse{}, err
	}

	logger.Info(string(custom_otel.NewLogMessage(
		span,
		logUUID.String(),
		owner,
		eventName,
		custom_otel.SetEventResource(dbPipeline),
	)))

	return resp, nil
}

func (h *PublicHandler) WatchPipeline(ctx context.Context, req *pipelinePB.WatchPipelineRequest) (*pipelinePB.WatchPipelineResponse, error) {

	eventName := "TriggerPipelineBinaryFileUpload"

	ctx, span := tracer.Start(ctx, eventName,
		trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	logUUID, _ := uuid.NewV4()

	logger, _ := logger.GetZapLogger(ctx)

	owner, err := resource.GetOwner(ctx, h.service.GetMgmtPrivateServiceClient(), h.service.GetRedisClient())
	if err != nil {
		span.SetStatus(1, err.Error())
		logger.Info(string(custom_otel.NewLogMessage(
			span,
			logUUID.String(),
			owner,
			eventName,
			custom_otel.SetErrorMessage(err.Error()),
		)))
		return &pipelinePB.WatchPipelineResponse{}, err
	}

	id, err := resource.GetRscNameID(req.GetName())
	if err != nil {
		span.SetStatus(1, err.Error())
		logger.Info(string(custom_otel.NewLogMessage(
			span,
			logUUID.String(),
			owner,
			eventName,
			custom_otel.SetErrorMessage(err.Error()),
			custom_otel.SetEventResource(req.GetName()),
		)))
		return &pipelinePB.WatchPipelineResponse{}, err
	}

	dbPipeline, err := h.service.GetPipelineByID(id, owner, true)
	if err != nil {
		span.SetStatus(1, err.Error())
		logger.Info(string(custom_otel.NewLogMessage(
			span,
			logUUID.String(),
			owner,
			eventName,
			custom_otel.SetErrorMessage(err.Error()),
			custom_otel.SetEventResource(req.GetName()),
		)))
		return &pipelinePB.WatchPipelineResponse{}, err
	}
	state, err := h.service.GetResourceState(dbPipeline.UID)
	if err != nil {
		span.SetStatus(1, err.Error())
		logger.Info(string(custom_otel.NewLogMessage(
			span,
			logUUID.String(),
			owner,
			eventName,
			custom_otel.SetErrorMessage(err.Error()),
			custom_otel.SetEventResource(req.GetName()),
		)))
		return &pipelinePB.WatchPipelineResponse{}, err
	}

	return &pipelinePB.WatchPipelineResponse{
		State: *state,
	}, nil
}

func (h *PublicHandler) GetOperation(ctx context.Context, req *pipelinePB.GetOperationRequest) (*pipelinePB.GetOperationResponse, error) {

	operationId, err := resource.GetOperationID(req.Name)
	if err != nil {
		return &pipelinePB.GetOperationResponse{}, err
	}
	operation, err := h.service.GetOperation(ctx, operationId)
	if err != nil {
		return &pipelinePB.GetOperationResponse{}, err
	}

	return &pipelinePB.GetOperationResponse{
		Operation: operation,
	}, nil
}
