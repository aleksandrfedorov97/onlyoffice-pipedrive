import React from "react";
import cx from "classnames";

type TitleProps = {
  text: string;
  large?: boolean;
};

export const OnlyofficeTitle: React.FC<TitleProps> = ({
  text,
  large = false,
}) => {
  const style = cx({
    "font-semibold text-slate-800 text-center": !!text,
    "text-base": large,
    "text-sm": !large,
  });

  return <p className={style}>{text}</p>;
};
