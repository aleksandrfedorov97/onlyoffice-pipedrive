import React from "react";
import md5 from "md5";
import { useSnapshot } from "valtio";
import { Command, Modal } from "@pipedrive/app-extensions-sdk";

import { useDeleteFile } from "@hooks/useDeleteFile";

import { isFileSupported } from "@utils/file";
import { getCurrentURL } from "@utils/url";

import { PipedriveSDK } from "@context/PipedriveContext";

import { File } from "src/types/file";

import Pencil from "@assets/pencil.svg";
import Trash from "@assets/trash.svg";

type FileActionsProps = {
  file: File;
};

export const OnlyofficeFileActions: React.FC<FileActionsProps> = ({ file }) => {
  const { sdk } = useSnapshot(PipedriveSDK);
  const { url, parameters } = getCurrentURL();
  const mutator = useDeleteFile(`${url}api/v1/files/${file.id}`);
  const handleDelete = () => {
    mutator
      .mutateAsync()
      .then(async () => {
        await sdk.execute(Command.SHOW_SNACKBAR, {
          message: `File ${file.name} has been removed`,
        });
      })
      .catch(async () => {
        await sdk.execute(Command.SHOW_SNACKBAR, {
          message: `Could not remove file ${file.name}`,
        });
      });
  };

  const handleEditor = async () => {
    await sdk.execute(Command.OPEN_MODAL, {
      type: Modal.CUSTOM_MODAL,
      action_id: process.env.PIPEDRIVE_EDITOR_MODAL_ID || "",
      data: {
        deal_id: parameters.get("selectedIds") || "",
        id: file.id,
        name: file.name,
        key: md5(file.id + file.update_time),
      },
    });
  };

  return (
    <>
      <div
        role="button"
        tabIndex={0}
        className={`${
          !isFileSupported(file.name)
            ? "hover:cursor-default opacity-50"
            : "hover:cursor-pointer"
        } mx-1`}
        onClick={() => handleEditor()}
        onKeyDown={() => handleEditor()}
      >
        <Pencil />
      </div>
      <div
        role="button"
        tabIndex={0}
        className="hover:cursor-pointer mx-1"
        onClick={() => handleDelete()}
        onKeyDown={() => handleDelete()}
      >
        <Trash />
      </div>
    </>
  );
};
