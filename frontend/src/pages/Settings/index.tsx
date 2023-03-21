import React, { useEffect, useState } from "react";
import { Command } from "@pipedrive/app-extensions-sdk";
import { useSnapshot } from "valtio";

import { OnlyofficeButton } from "@components/button";
import { OnlyofficeInput } from "@components/input";
import { OnlyofficeTitle } from "@components/title";
import { OnlyofficeSpinner } from "@components/spinner";
import { OnlyofficeBackgroundError } from "@layouts/ErrorBackground";

import { postSettings, getSettings } from "@services/settings";
import { getPipedriveMe } from "@services/me";

import { PipedriveSDK } from "@context/PipedriveContext";
import { AuthToken } from "@context/TokenContext";

export const SettingsPage: React.FC = () => {
  const { sdk } = useSnapshot(PipedriveSDK);
  const { access_token: accessToken, error } = useSnapshot(AuthToken);
  const [admin, setAdmin] = useState(false);
  const [loading, setLoading] = useState(true);
  const [address, setAddress] = useState<string | undefined>(undefined);
  const [secret, setSecret] = useState<string | undefined>(undefined);
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    if (accessToken && !error) {
      getPipedriveMe(
        `${window.parent[0].location.ancestorOrigins[0]}/api/v1/users/me`
      )
        .then(async (ures) => {
          try {
            if (ures.data.access.find((a) => a.app === "global" && a.admin)) {
              const res = await getSettings(sdk);
              setAddress(res.doc_address);
              setSecret(res.doc_secret);
              setAdmin(true);
            }
          } catch {
            setAdmin(false);
          } finally {
            setLoading(false);
          }
        })
        .catch(() => {
          setLoading(false);
        });
    }

    if (error) setLoading(false);
  }, [sdk, accessToken, error]);

  const handleSettings = async () => {
    if (address && secret) {
      try {
        setSaving(true);
        await postSettings(sdk, address, secret);
        await sdk.execute(Command.SHOW_SNACKBAR, {
          message: "ONLYOFFICE settings have been saved",
        });
      } catch {
        await sdk.execute(Command.SHOW_SNACKBAR, {
          message: "Could not save ONLYOFFICE settings",
        });
      } finally {
        setSaving(false);
      }
    }
  };

  return (
    <div className="w-screen h-screen">
      {loading && !error && (
        <div className="h-full w-full flex justify-center items-center">
          <OnlyofficeSpinner />
        </div>
      )}
      {!loading && error && (
        <OnlyofficeBackgroundError
          title="Error"
          subtitle="Could not fetch plugin settings. Something went wrong. Please reload the pipedrive window"
          button="Reload"
          onClick={() => window.location.reload()}
        />
      )}
      {!loading && !error && !admin && (
        <OnlyofficeBackgroundError
          title="Access Denied"
          subtitle="Something went wrong or access denied"
          button="Reload"
          onClick={() => window.location.reload()}
        />
      )}
      {!loading && !error && admin && (
        <>
          <div className="flex flex-col items-start pl-5 pr-5 pt-12 pb-7">
            <div className="pb-2">
              <OnlyofficeTitle text="Configure ONLYOFFICE app settings" />
            </div>
            <p className="text-slate-800 font-normal text-base text-left">
              The plugin which enables the users to edit office documents from
              Pipedrive using ONLYOFFICE Document Server, allows multiple users
              to collaborate in real time and to save back those changes to
              Pipedrive
            </p>
          </div>
          <div className="max-w-[320px]">
            <div className="pl-5 pr-5 pb-5">
              <OnlyofficeInput
                text="Document Server Address"
                valid={!!address}
                disabled={saving}
                value={address}
                onChange={(e) => setAddress(e.target.value)}
              />
            </div>
            <div className="pl-5 pr-5">
              <OnlyofficeInput
                text="Document Server Secret"
                valid={!!secret}
                disabled={saving}
                value={secret}
                onChange={(e) => setSecret(e.target.value)}
                type="password"
              />
            </div>
            <div className="flex justify-start items-center mt-8 ml-5">
              <OnlyofficeButton
                text="Save"
                primary
                disabled={saving}
                onClick={handleSettings}
              />
            </div>
          </div>
        </>
      )}
    </div>
  );
};
