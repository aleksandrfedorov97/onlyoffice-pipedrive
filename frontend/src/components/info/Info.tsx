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
