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

import Docx from "@assets/docx.svg";
import Pptx from "@assets/pptx.svg";
import Xlsx from "@assets/xlsx.svg";
import Unsupported from "@assets/unsupported.svg";
import Supported from "@assets/supported.svg";

import wordIcon from "@assets/word.ico";
import slideIcon from "@assets/slide.ico";
import cellIcon from "@assets/cell.ico";
import genericIcon from "@assets/generic.ico";

const DOCUMENT_EXTS = [
  "doc",
  "docx",
  "docm",
  "dot",
  "dotx",
  "dotm",
  "odt",
  "fodt",
  "ott",
  "rtf",
  "txt",
  "html",
  "htm",
  "mht",
  "xml",
  "pdf",
  "djvu",
  "fb2",
  "epub",
  "xps",
  "oxps",
];

const SPREADSHEET_EXTS = [
  "xls",
  "xlsx",
  "xlsm",
  "xlt",
  "xltx",
  "xltm",
  "ods",
  "fods",
  "ots",
  "csv",
];

const PRESENTATION_EXTS = [
  "pps",
  "ppsx",
  "ppsm",
  "ppt",
  "pptx",
  "pptm",
  "pot",
  "potx",
  "potm",
  "odp",
  "fodp",
  "otp",
];

const EDITABLE_EXTS = ["docx", "pptx", "xlsx"];
const OPENABLE_EXTS =
  DOCUMENT_EXTS.concat(SPREADSHEET_EXTS).concat(PRESENTATION_EXTS);

const WORD = "word";
const SLIDE = "slide";
const CELL = "cell";

export const getFileParts = (filename: string): [string, string] => {
  const [name, ext] = filename.split(".");
  return [name, ext];
};

const getFileExt = (filename: string): string =>
  filename.split(".").pop() || "";

export const isFileEditable = (filename: string) => {
  const ext = getFileExt(filename).toLowerCase();
  return EDITABLE_EXTS.includes(ext);
};

export const isFileSupported = (filename: string) => {
  const e = getFileExt(filename).toLowerCase();
  return OPENABLE_EXTS.includes(e);
};

export const getFileType = (filename: string) => {
  const e = getFileExt(filename).toLowerCase();

  if (DOCUMENT_EXTS.includes(e)) return WORD;
  if (SPREADSHEET_EXTS.includes(e)) return CELL;
  if (PRESENTATION_EXTS.includes(e)) return SLIDE;

  return null;
};

export const getMimeType = (filename: string) => {
  const e = getFileExt(filename).toLowerCase();

  switch (e) {
    case "docx":
      return "application/vnd.openxmlformats-officedocument.wordprocessingml.document";
    case "pptx":
      return "application/vnd.openxmlformats-officedocument.presentationml.presentation";
    case "xlsx":
      return "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet";
    default:
      return "application/octet-stream";
  }
};

export const getFileIcon = (filename: string) => {
  const e = getFileExt(filename).toLowerCase();

  if (e === "docx") return Docx;
  if (e === "xlsx") return Xlsx;
  if (e === "pptx") return Pptx;
  if (
    DOCUMENT_EXTS.includes(e) ||
    SPREADSHEET_EXTS.includes(e) ||
    PRESENTATION_EXTS.includes(e)
  )
    return Supported;

  return Unsupported;
};

export const getCreateFileUrl = (
  fileType: "docx" | "pptx" | "xlsx" | undefined
) => {
  switch (fileType) {
    case "docx":
      return encodeURIComponent(process.env.WORD_FILE || "");
    case "pptx":
      return encodeURIComponent(process.env.SLIDE_FILE || "");
    case "xlsx":
      return encodeURIComponent(process.env.SPREADSHEET_FILE || "");
    default:
      return encodeURIComponent(process.env.WORD_FILE || "");
  }
};

export const formatBytes = (bytes: number, decimals = 2) => {
  if (!+bytes) return "0 Bytes";

  const k = 1024;
  const dm = decimals < 0 ? 0 : decimals;
  const sizes = ["Bytes", "KB", "MB", "GB", "TB", "PB", "EB", "ZB", "YB"];

  const i = Math.floor(Math.log(bytes) / Math.log(k));

  return `${parseFloat((bytes / k ** i).toFixed(dm))} ${sizes[i]}`;
};

export const getFileFavicon = (filename: string) => {
  const e = getFileExt(filename).toLowerCase();
  if (DOCUMENT_EXTS.includes(e)) return wordIcon;
  if (PRESENTATION_EXTS.includes(e)) return slideIcon;
  if (SPREADSHEET_EXTS.includes(e)) return cellIcon;

  return genericIcon;
};
