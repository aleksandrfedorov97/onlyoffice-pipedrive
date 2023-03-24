import React from "react";
import { useSearchParams } from "react-router-dom";
import { useTranslation } from "react-i18next";
import { DocumentEditor } from "@onlyoffice/document-editor-react";

import { OnlyofficeButton } from "@components/button";
import { OnlyofficeError } from "@components/error";
import { OnlyofficeSpinner } from "@components/spinner";

import { useBuildConfig } from "@hooks/useBuildConfig";

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
  const { t } = useTranslation();
  const [params] = useSearchParams();
  const { isLoading, error, data } = useBuildConfig(
    params.get("token") || "",
    params.get("id") || "",
    params.get("name") || "new.docx",
    params.get("key") || new Date().toTimeString(),
    params.get("deal_id") || "1"
  );

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
              text={t("button.cancel", "Cancel")}
              fullWidth
              onClick={() => window.close()}
            />
          </div>
        </div>
      )}
      {!!error && (
        <div className="w-full h-full flex justify-center flex-col items-center mb-1">
          <Icon />
          <OnlyofficeError
            text={t(
              "editor.error",
              "Could not open the file. Something went wrong"
            )}
          />
          <div className="pt-5">
            <OnlyofficeButton
              primary
              text={t("button.close", "Close")}
              onClick={() => window.close()}
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
              },
              token: data.token,
              type: data.type,
              events: {
                onRequestClose: async () => {
                  window.close();
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
