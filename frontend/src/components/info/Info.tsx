import React from "react";

import { OnlyofficeSubtitle, OnlyofficeTitle } from "@components/title";

type FileInfoProps = {
  info: {
    [key: string]: string;
  };
};

export const OnlyofficeFileInfo: React.FC<FileInfoProps> = ({ info }) => (
  <table className="table-auto mx-1">
    <tbody>
      {Object.keys(info).map((subtitle) => (
        <tr key={subtitle + info[subtitle]} className="text-left">
          <td className="flex justify-start mr-10 my-1">
            <OnlyofficeSubtitle text={subtitle} />
          </td>
          <td>
            <div className="flex justify-start">
              <OnlyofficeTitle text={info[subtitle]} />
            </div>
          </td>
        </tr>
      ))}
    </tbody>
  </table>
);
