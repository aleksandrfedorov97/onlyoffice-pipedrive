import AppExtensionsSDK from "@pipedrive/app-extensions-sdk";
import { proxy } from "valtio";

export const PipedriveSDK = proxy({ sdk: new AppExtensionsSDK().initialize() });
