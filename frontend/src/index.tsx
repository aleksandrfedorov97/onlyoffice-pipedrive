import React from "react";
import ReactDOM from "react-dom/client";
import { QueryClient, QueryClientProvider } from "react-query";

import { TokenProvider } from "@context/TokenContext";

import App from "./App";
// import "./i18n";

import "@assets/index.css";

const root = ReactDOM.createRoot(
  document.getElementById("root") as HTMLElement
);

const queryClient = new QueryClient();
root.render(
  <React.StrictMode>
    <TokenProvider>
      <QueryClientProvider client={queryClient}>
        <App />
      </QueryClientProvider>
    </TokenProvider>
  </React.StrictMode>
);
