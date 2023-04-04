import { uuidv4 } from 'https://jslib.k6.io/k6-utils/1.4.0/index.js';

let proto
let pHost, cHost, mHost
let pPrivatePort, pPublicPort, cPublicPort, mPublicPort

if (__ENV.MODE == "api-gateway") {
  // api-gateway mode
  proto = "http"
  pHost = cHost = mHost = "api-gateway"
  pPrivatePort = 3081
  pPublicPort = cPublicPort = mPublicPort = 8080
} else if (__ENV.MODE == "localhost") {
  // localhost mode for GitHub Actions
  proto = "http"
  pHost = cHost = mHost = "localhost"
  pPrivatePort = 3081
  pPublicPort = cPublicPort = mPublicPort = 8080
} else {
  // direct microservice mode
  proto = "http"
  pHost = "pipeline-backend"
  cHost = "connector-backend"
  mHost = "model-backend"
  pPrivatePort = 3081
  pPublicPort = 8081
  cPublicPort = 8082
  mPublicPort = 8083
}

export const pipelinePrivateHost = `${proto}://${pHost}:${pPrivatePort}`;
export const pipelinePublicHost = `${proto}://${pHost}:${pPublicPort}`;
export const pipelineGRPCPrivateHost = `${pHost}:${pPrivatePort}`;
export const pipelineGRPCPublicHost = `${pHost}:${pPublicPort}`;
export const connectorPublicHost = `${proto}://${cHost}:${cPublicPort}`;
export const connectorGRPCPublicHost = `${cHost}:${cPublicPort}`;
export const modelPublicHost = `${proto}://${mHost}:${mPublicPort}`;

export const dogImg = open(`${__ENV.TEST_FOLDER_ABS_PATH}/integration-test/data/dog.jpg`, "b");
export const catImg = open(`${__ENV.TEST_FOLDER_ABS_PATH}/integration-test/data/cat.jpg`, "b");
export const bearImg = open(`${__ENV.TEST_FOLDER_ABS_PATH}/integration-test/data/bear.jpg`, "b");
export const dogRGBAImg = open(`${__ENV.TEST_FOLDER_ABS_PATH}/integration-test/data/dog-rgba.png`, "b");

export const det_model = open(`${__ENV.TEST_FOLDER_ABS_PATH}/integration-test/data/dummy-det-model.zip`, "b");
export const model_def_name = "model-definitions/local"
export const model_id = "dummy-det"

export const params = {
  headers: {
    "Content-Type": "application/json",
  },
};

const randomUUID = uuidv4();
export const paramsWithJwt = {
  headers: {
    "Content-Type": "application/json",
    "Jwt-Sub": randomUUID,
  },
}

export const detSyncHTTPSingleModelRecipe = {
  recipe: {
    source: "source-connectors/source-http",
    models: [
      `models/${model_id}`,
    ],
    destination: "destination-connectors/destination-http"
  },
};

export const detSyncGRPCSingleModelRecipe = {
  recipe: {
    source: "source-connectors/source-grpc",
    models: [
      `models/${model_id}`,
    ],
    destination: "destination-connectors/destination-grpc"
  },
};

export const detSyncHTTPMultiModelRecipe = {
  recipe: {
    source: "source-connectors/source-http",
    models: [
      `models/${model_id}`,
      `models/${model_id}`,
    ],
    destination: "destination-connectors/destination-http"
  },
};

export const detSynGRPCMultiModelRecipe = {
  recipe: {
    source: "source-connectors/source-grpc",
    models: [
      `models/${model_id}`,
      `models/${model_id}`,
    ],
    destination: "destination-connectors/destination-grpc"
  },
};

export const dstCSVConnID = "some-cool-name-for-dst-csv-connector"

export const detAsyncSingleModelRecipe = {
  recipe: {
    source: "source-connectors/source-http",
    models: [
      `models/${model_id}`
    ],
    destination: `destination-connectors/${dstCSVConnID}`
  },
};

export const detAsyncMultiModelRecipe = {
  recipe: {
    source: "source-connectors/source-http",
    models: [
      `models/${model_id}`,
      `models/${model_id}`,
    ],
    destination: `destination-connectors/${dstCSVConnID}`
  },
};
