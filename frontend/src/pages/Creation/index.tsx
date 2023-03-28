import React, { useEffect, useState } from "react";
import AppExtensionsSDK from "@pipedrive/app-extensions-sdk";
import { Tabs, TabList, Tab, TabPanel } from "react-tabs";
import { useTranslation } from "react-i18next";

import { Creation } from "./Creation";
import { Upload } from "./Upload";

export const CreatePage: React.FC = () => {
  const { t } = useTranslation();
  const [selected, setSelected] = useState(0);
  useEffect(() => {
    new AppExtensionsSDK().initialize({
      size: {
        height: 500,
        width: 622,
      },
    });
  });

  return (
    <div className="relative w-full h-full flex flex-col overflow-hidden">
      <Tabs
        className="flex justify-center items-start flex-col h-screen"
        onSelect={(index) => setSelected(index)}
      >
        <TabList
          className="flex justify-start items-center min-h-[40px] w-full"
          style={{ backgroundColor: "#F7F7F7" }}
        >
          <Tab
            id="create-file"
            className={`flex justify-center items-center text-sm font-inter outline-none hover:cursor-pointer min-h-[40px] ${
              selected === 0
                ? "text-blue-600 border-b-2 border-blue-600"
                : "text-slate-500"
            }`}
            style={{ margin: "0 0 0 1em", padding: "1em" }}
          >
            {t("button.creation.create", "Create")}
          </Tab>
          <Tab
            id="upload-file"
            className={`flex justify-center items-center text-sm font-inter outline-none hover:cursor-pointer min-h-[40px] ${
              selected === 1
                ? "text-blue-600 border-b-2 border-blue-600"
                : "text-slate-500"
            }`}
            style={{ margin: "0 0 0 1em", padding: "1em" }}
          >
            {t("button.creation.upload", "Upload")}
          </Tab>
        </TabList>
        <TabPanel
          className={`${
            selected === 0 ? "h-[calc(100%-40px)] w-full" : "h-0 w-0"
          }`}
        >
          <Creation />
        </TabPanel>
        <TabPanel
          className={`${
            selected === 1 ? "h-[calc(100%-40px)] w-full" : "h-0 w-0"
          }`}
        >
          <Upload />
        </TabPanel>
      </Tabs>
    </div>
  );
};
