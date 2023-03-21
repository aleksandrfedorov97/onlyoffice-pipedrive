import React from "react";

import { OnlyofficeButton } from "@components/button";
import { OnlyofficeSubtitle } from "@components/title";
import { OnlyofficeError } from "@components/error/Error";

import BackgroundError from "@assets/background-error.svg";

type ErrorProps = {
  title: string;
  subtitle: string;
  button?: string;
  onClick?: React.MouseEventHandler<HTMLButtonElement> | undefined;
};

export const OnlyofficeBackgroundError: React.FC<ErrorProps> = ({
  title,
  subtitle,
  button,
  onClick,
}) => (
  <div className="w-full h-full flex justify-center flex-col items-center overflow-hidden">
    <div className="absolute flex justify-center items-center h-full overflow-hidden">
      <BackgroundError />
    </div>
    <div className="pb-5">
      <OnlyofficeError text={title} />
    </div>
    <OnlyofficeSubtitle text={subtitle} />
    {onClick && button && (
      <div className="pt-5 z-[100]">
        <OnlyofficeButton primary text={button} onClick={onClick} />
      </div>
    )}
  </div>
);
