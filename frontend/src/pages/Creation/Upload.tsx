import { FileRejection, DropEvent } from "react-dropzone";
import React, { useEffect, useState } from "react";
import AppExtensionsSDK, { Command } from "@pipedrive/app-extensions-sdk";
import { useTranslation } from "react-i18next";

import { OnlyofficeButton } from "@components/button";
import { OnlyofficeDragDrop } from "@components/drop";

import { uploadFile } from "@services/file";

import { getCurrentURL } from "@utils/url";

const onDrop = <T extends File>(
  acceptedFiles: T[],
  _: FileRejection[],
  __: DropEvent
): Promise<void> => {
  const { url, parameters } = getCurrentURL();
  return uploadFile(
    `${url}api/v1/files`,
    parameters.get("selectedIds") || "",
    acceptedFiles[0]
  );
};

export const Upload: React.FC = () => {
  const { t } = useTranslation();
  const [sdk, setSDK] = useState<AppExtensionsSDK | null>();
  useEffect(() => {
    new AppExtensionsSDK()
      .initialize()
      .then((s) => setSDK(s))
      .catch(() => setSDK(null));
  }, []);

  return (
    <div className="h-full">
      <div className="h-[calc(100%-3rem)] overflow-hidden">
        <div className="px-5 py-20 flex flex-col justify-center items-start h-full">
          <OnlyofficeDragDrop
            errorText={
              t(
                "upload.error",
                "Could not upload your file. Please contact ONLYOFFICE support."
              ) ||
              "Could not upload your file. Please contact ONLYOFFICE support."
            }
            uploadingText={
              t("upload.uploading", "Uploading...") || "Uploading..."
            }
            selectText={t("upload.select", "Select a file") || "Select a file"}
            dragdropText={
              t("upload.dragdrop", "or drag and drop here") ||
              "or drag and drop here"
            }
            subtext={
              t("upload.subtext", "File size is limited") ||
              "File size is limited"
            }
            onDrop={onDrop}
          />
        </div>
      </div>
      <div className="h-[48px] flex items-center w-full">
        <div className="flex justify-between items-center w-full">
          <div className="mx-5">
            <OnlyofficeButton
              text={t("button.cancel", "Cancel")}
              onClick={async () => {
                await sdk?.execute(Command.CLOSE_MODAL);
              }}
            />
          </div>
        </div>
      </div>
    </div>
  );
};
