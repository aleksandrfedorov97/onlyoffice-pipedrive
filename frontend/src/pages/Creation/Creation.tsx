import React, { useEffect, useState } from "react";
import md5 from "md5";
import AppExtensionsSDK, { Command } from "@pipedrive/app-extensions-sdk";

import { OnlyofficeButton } from "@components/button";
import { OnlyofficeInput } from "@components/input";
import { OnlyofficeTile } from "@components/tile";
import { OnlyofficeTitle } from "@components/title";

import { uploadFile } from "@services/file";

import { getFileIcon, getMimeType } from "@utils/file";
import { getCurrentURL } from "@utils/url";

export const Creation: React.FC = () => {
  const [sdk, setSDK] = useState<AppExtensionsSDK | null>();
  const [file, setFile] = useState("New Document");
  const [fileType, setFileType] = useState<"docx" | "pptx" | "xlsx">("docx");
  const handleChangeFile = (newType: "docx" | "pptx" | "xlsx") => {
    setFileType(newType);
  };

  useEffect(() => {
    new AppExtensionsSDK()
      .initialize()
      .then((s) => setSDK(s))
      .catch(() => setSDK(null));
  }, []);

  return (
    <div className="h-full">
      <div className="h-[calc(100%-3rem)] overflow-scroll">
        <div className="px-5 flex flex-col justify-center items-start h-full">
          <OnlyofficeTitle text="Create with ONLYOFFICE" large />
          <div className="pt-4 w-[330px]">
            <OnlyofficeInput
              text="Title"
              labelSize="sm"
              value={file}
              onChange={(e) => setFile(e.target.value)}
            />
          </div>
          <div className="flex justify-center items-center pt-10">
            <div>
              <OnlyofficeTile
                Icon={getFileIcon("sample.docx")}
                text="Document"
                onClick={() => handleChangeFile("docx")}
                onKeyDown={() => handleChangeFile("docx")}
                selected={fileType === "docx"}
              />
            </div>
            <div className="px-1">
              <OnlyofficeTile
                Icon={getFileIcon("sample.xlsx")}
                text="Spreadsheet"
                onClick={() => handleChangeFile("xlsx")}
                onKeyDown={() => handleChangeFile("xlsx")}
                selected={fileType === "xlsx"}
              />
            </div>
            <div>
              <OnlyofficeTile
                Icon={getFileIcon("sample.pptx")}
                text="Presentation"
                onClick={() => handleChangeFile("pptx")}
                onKeyDown={() => handleChangeFile("pptx")}
                selected={fileType === "pptx"}
              />
            </div>
          </div>
        </div>
      </div>
      <div
        className="h-[48px] flex items-center w-full"
        style={{ backgroundColor: "#F7F7F7" }}
      >
        <div className="flex justify-between items-center w-full">
          <div className="mx-5">
            <OnlyofficeButton
              text="Cancel"
              onClick={async () => {
                await sdk?.execute(Command.CLOSE_MODAL);
              }}
            />
          </div>
          <div className="mx-5">
            <OnlyofficeButton
              text="Create document"
              primary
              onClick={async () => {
                const token = await sdk?.execute(Command.GET_SIGNED_TOKEN);
                if (!token) return;
                const { url, parameters } = getCurrentURL();
                const filename = `${file}.${fileType}`;
                const binary = new File([], filename, {
                  type: getMimeType(filename),
                });
                try {
                  const res = await uploadFile(
                    `${url}api/v1/files`,
                    parameters.get("selectedIds") || "",
                    binary
                  );
                  window.open(
                    `/editor?token=${token.token}&id=${res.data.id}&deal_id=${
                      res.data.deal_id
                    }&name=${res.data.name}&key=${md5(
                      res.data.id + res.data.update_time
                    )}`
                  );
                } catch {
                  await sdk?.execute(Command.SHOW_SNACKBAR, {
                    message: "Could not create a new file",
                  });
                }
              }}
            />
          </div>
        </div>
      </div>
    </div>
  );
};
