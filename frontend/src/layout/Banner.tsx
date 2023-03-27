import React from "react";

import { OnlyofficeTitle, OnlyofficeSubtitle } from "@components/title";

import BannerIcon from "@assets/banner.svg";
import { useTranslation } from "react-i18next";

export const Banner: React.FC = () => {
  const { t } = useTranslation();
  return (
    <div
      className="flex justify-between items-center p-5 mt-5 mb-5"
      style={{ backgroundColor: "#F6F6F6", border: "1px solid #EFEFEF" }}
    >
      <div className="w-2/12">
        <BannerIcon />
      </div>
      <div className="w-7/12 flex justify-center items-start">
        <div className="flex justify-start items-center flex-col cursor-default ml-5 mr-5">
          <div className="w-full h-1/2 flex">
            <OnlyofficeTitle
              text={
                t("banner.title", "ONLYOFFICE Docs Cloud") ||
                "ONLYOFFICE Docs Cloud"
              }
              large
            />
          </div>
          <div className="w-full h-[40px] overflow-hidden">
            <OnlyofficeSubtitle
              text={
                t(
                  "banner.subtitle",
                  "Easily launch the editors in the cloud without downloading and installation"
                ) ||
                "Easily launch the editors in the cloud without downloading and installation"
              }
              center={false}
            />
          </div>
        </div>
      </div>
      <div className="w-3/12">
        <button
          type="button"
          className="pl-5 pr-5 pt-2 pb-2 text-sm whitespace-nowrap rounded overflow-hidden text-ellipsis inline-block max-w-[200px] cursor-pointer hover:shadow-sm duration-200"
          style={{ backgroundColor: "#192435", color: "#FFFFFF" }}
          onClick={() =>
            window.open(
              "https://www.onlyoffice.com/docs-registration.aspx?referer=pipedrive"
            )
          }
        >
          {t("button.getnow", "Get Now")}
        </button>
      </div>
    </div>
  );
};
