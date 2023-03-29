/**
 *
 * (c) Copyright Ascensio System SIA 2023
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

import React, { useEffect, useState } from "react";
import md5 from "md5";
import AppExtensionsSDK, { Command } from "@pipedrive/app-extensions-sdk";
import { useTranslation } from "react-i18next";

import { OnlyofficeButton } from "@components/button";
import { OnlyofficeInput } from "@components/input";
import { OnlyofficeTile } from "@components/tile";
import { OnlyofficeTitle } from "@components/title";

import { uploadFile } from "@services/file";

import { getFileParts, getFileIcon, getMimeType } from "@utils/file";
import { getCurrentURL } from "@utils/url";

import Redirect from "@assets/redirect.svg";

export const Creation: React.FC = () => {
  const { t } = useTranslation();
  const [sdk, setSDK] = useState<AppExtensionsSDK | null>();
  const [file, setFile] = useState(
    t("document.new", "New Document") || "New Document"
  );
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
    <div className="h-full w-full">
      <div className="h-[calc(100%-3rem)] w-full overflow-hidden">
        <div className="px-5 flex flex-col justify-center items-start h-full">
          <OnlyofficeTitle
            text={t("creation.title", "Create with ONLYOFFICE")}
            large
          />
          <div className="w-full pt-4">
            <OnlyofficeInput
              text={t("creation.inputs.title", "Title")}
              labelSize="sm"
              valid={file.length <= 190}
              errorText={
                t(
                  "creation.inputs.error",
                  "File name length should be less than 190 characters"
                ) || "File name length should be less than 190 characters"
              }
              value={file}
              onChange={(e) => setFile(e.target.value)}
            />
          </div>
          <div className="w-full flex pt-5">
            <div className="grow">
              <OnlyofficeTile
                Icon={getFileIcon("sample.docx")}
                text={t("creation.tiles.doc", "Document")}
                onClick={() => handleChangeFile("docx")}
                onKeyDown={() => handleChangeFile("docx")}
                selected={fileType === "docx"}
              />
            </div>
            <div className="grow px-5">
              <OnlyofficeTile
                Icon={getFileIcon("sample.xlsx")}
                text={t("creation.tiles.spreadsheet", "Spreadsheet")}
                onClick={() => handleChangeFile("xlsx")}
                onKeyDown={() => handleChangeFile("xlsx")}
                selected={fileType === "xlsx"}
              />
            </div>
            <div className="grow">
              <OnlyofficeTile
                Icon={getFileIcon("sample.pptx")}
                text={t("creation.tiles.presentation", "Presentation")}
                onClick={() => handleChangeFile("pptx")}
                onKeyDown={() => handleChangeFile("pptx")}
                selected={fileType === "pptx"}
              />
            </div>
          </div>
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
          <div className="mx-5">
            <OnlyofficeButton
              disabled={file.length > 190}
              text={t("button.create", "Create document")}
              primary
              Icon={<Redirect />}
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
                  const [name, ext] = getFileParts(res.data.name);
                  window.open(
                    `/editor?token=${token.token}&id=${res.data.id}&deal_id=${
                      res.data.deal_id
                    }&name=${`${encodeURIComponent(
                      name.substring(0, 190)
                    )}.${ext}`}&key=${md5(res.data.id + res.data.update_time)}`
                  );
                } catch {
                  await sdk?.execute(Command.SHOW_SNACKBAR, {
                    message: t("creation.error", "Could not create a new file"),
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
