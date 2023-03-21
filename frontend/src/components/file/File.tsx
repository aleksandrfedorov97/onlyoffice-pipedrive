import React, { useState } from "react";

import DetailsIcon from "@assets/arrow-down.svg";

type FileProps = {
  Icon: any;
  name: string;
  supported?: boolean;
  actions?: React.ReactNode;
  children?: React.ReactNode;
  onClick?: React.MouseEventHandler<HTMLButtonElement>;
};

export const OnlyofficeFile: React.FC<FileProps> = ({
  Icon,
  name,
  supported = false,
  actions,
  children,
  onClick,
}) => {
  const [showDetails, setShowDetails] = useState(false);
  return (
    <>
      <div className="flex items-center w-full border-b py-2 my-1">
        <div className="flex items-center justify-center">
          <div
            role="button"
            tabIndex={0}
            onClick={() => setShowDetails(!showDetails)}
            onKeyDown={() => setShowDetails(!showDetails)}
            className={`w-[16px] h-[16px] hover:cursor-pointer mx-1 ${
              showDetails ? "rotate-180" : "rotate-0"
            }`}
          >
            <DetailsIcon />
          </div>
        </div>
        <div className="flex items-center justify-start w-3/4">
          <div className="w-[32px] h-[32px]">
            <Icon />
          </div>
          <button
            className={`${
              supported && onClick ? "cursor-pointer" : "cursor-default"
            } text-left font-semibold font-sans md:text-sm text-xs px-2 w-full h-[32px] overflow-hidden text-ellipsis whitespace-nowrap`}
            type="button"
            onClick={onClick}
          >
            {name}
          </button>
        </div>
        <div className="flex items-center justify-end w-1/4">{actions}</div>
      </div>
      <div
        className={`overflow-hidden transition-all ${
          showDetails ? "h-[200px]" : "h-[0px]"
        }`}
      >
        {children}
      </div>
    </>
  );
};
