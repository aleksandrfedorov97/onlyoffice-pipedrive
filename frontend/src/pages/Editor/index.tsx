import React, { useEffect } from "react";
import { useSearchParams } from "react-router-dom";
import { useTranslation } from "react-i18next";
import { useSnapshot } from "valtio";
import { Command } from "@pipedrive/app-extensions-sdk";
import { DocumentEditor } from "@onlyoffice/document-editor-react";

import { OnlyofficeButton } from "@components/button";
import { OnlyofficeError } from "@components/error";
import { OnlyofficeSpinner } from "@components/spinner";

import { useBuildConfig } from "@hooks/useBuildConfig";
import { PipedriveSDK } from "@context/PipedriveContext";

import Icon from "@assets/nofile.svg";

const onEditor = () => {
  const loader = document.getElementById("eloader");
  if (loader) {
    loader.classList.add("opacity-0");
    loader.classList.add("-z-100");
    loader.classList.add("hidden");
  }

  const editor = document.getElementById("editor");
  if (editor) {
    editor.classList.remove("opacity-0");
  }
};

export const OnlyofficeEditorPage: React.FC = () => {
  const [params] = useSearchParams();
  const pData = JSON.parse(params.get("data") || "{}");
  const { sdk } = useSnapshot(PipedriveSDK);
  const { t } = useTranslation();
  const { isLoading, error, data } = useBuildConfig(
    pData.id || "",
    pData.name || "new.docx",
    pData.key || new Date().toString(),
    pData.deal_id || "1"
  );

  useEffect(() => {
    if (sdk) {
      (async () => {
        await sdk.execute(Command.RESIZE, {
          height: 500,
          width: 700,
        });
      })();
    }
  }, [sdk]);

  const validConfig = !error && !isLoading && data;
  return (
    <div className="w-full h-full">
      {!error && (
        <div
          id="eloader"
          className="relative w-full h-full flex flex-col small:justify-between justify-center items-center transition duration-250 ease-linear"
        >
          <div className="pb-5 small:h-full small:flex small:items-center">
            <OnlyofficeSpinner />
          </div>
          <div className="small:mb-5 small:px-5 small:w-full">
            <OnlyofficeButton
              primary
              text="Cancel"
              fullWidth
              onClick={() => sdk.execute(Command.CLOSE_MODAL)}
            />
          </div>
        </div>
      )}
      {!!error && (
        <div className="w-full h-full flex justify-center flex-col items-center mb-1">
          <Icon />
          <OnlyofficeError
            text={
              t("editor.error") ||
              "Could not open the file. Something went wrong"
            }
          />
          <div className="pt-5">
            <OnlyofficeButton
              primary
              text={t("button.back") || "Go back"}
              onClick={() => sdk.execute(Command.CLOSE_MODAL)}
            />
          </div>
        </div>
      )}
      {validConfig && data.server_url && (
        <div
          id="editor"
          className="w-full h-full opacity-0 transition duration-250 ease-linear"
        >
          <DocumentEditor
            id="docxEditor"
            documentServerUrl={data.server_url}
            config={{
              document: {
                fileType: data.document.fileType,
                key: data.document.key,
                title: data.document.title,
                url: data.document.url,
                permissions: data.document.permissions,
              },
              documentType: data.documentType,
              editorConfig: {
                callbackUrl: data.editorConfig.callbackUrl,
                user: data.editorConfig.user,
                lang: data.editorConfig.lang,
                customization: {
                  goback: {
                    requestClose: true,
                    text: "Close",
                  },
                },
              },
              token: data.token,
              type: data.type,
              events: {
                onRequestClose: async () => {
                  await sdk.execute(Command.CLOSE_MODAL);
                },
                onAppReady: onEditor,
                onError: () => {
                  onEditor();
                },
                onWarning: onEditor,
              },
            }}
          />
        </div>
      )}
    </div>
  );
};

export default OnlyofficeEditorPage;
