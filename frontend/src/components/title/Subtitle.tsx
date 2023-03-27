import React from "react";
import cx from "classnames";

type SubtitleProps = {
  text: string;
  large?: boolean;
  center?: boolean;
};

export const OnlyofficeSubtitle: React.FC<SubtitleProps> = ({
  text,
  large = false,
  center = true,
}) => {
  const style = cx({
    "text-slate-700 font-normal": !!text,
    "text-center": center,
    "text-sm": !large,
    "text-base": large,
  });

  return <p className={style}>{text}</p>;
};
