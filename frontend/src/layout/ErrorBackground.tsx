import React from "react";

import { OnlyofficeButton } from "@components/button";
import { OnlyofficeSubtitle } from "@components/title";
import { OnlyofficeError } from "@components/error/Error";

import BackgroundError from "@assets/background-error.svg";

type ErrorProps = {
  Icon: any;
  title: string;
  subtitle: string;
  button?: string;
  onClick?: React.MouseEventHandler<HTMLButtonElement> | undefined;
};

export const OnlyofficeBackgroundError: React.FC<ErrorProps> = ({
  Icon,
  title,
  subtitle,
  button,
  onClick,
}) => (
  <div className="w-full h-full flex justify-center flex-col items-center overflow-hidden">
    <div className="flex justify-center items-center overflow-hidden">
      {Icon}
    </div>
    <div>
      <OnlyofficeError text={title} />
    </div>
    <div className="w-1/2 pt-2">
      <OnlyofficeSubtitle text={subtitle} />
    </div>
    {onClick && button && (
      <div className="pt-5 z-[100]">
        <OnlyofficeButton primary text={button} onClick={onClick} />
      </div>
    )}
  </div>
);
