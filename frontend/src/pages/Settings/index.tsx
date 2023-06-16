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

import React, { useEffect, useState } from "react";
import AppExtensionsSDK, { Command } from "@pipedrive/app-extensions-sdk";
import { useSnapshot } from "valtio";
import { useTranslation } from "react-i18next";

import { OnlyofficeButton } from "@components/button";
import { OnlyofficeInput } from "@components/input";
import { OnlyofficeTitle } from "@components/title";
import { OnlyofficeSpinner } from "@components/spinner";
import { OnlyofficeBackgroundError } from "@layouts/ErrorBackground";
import { Banner } from "@layouts/Banner";

import { postSettings, getSettings } from "@services/settings";
import { getPipedriveMe } from "@services/me";

import { AuthToken } from "@context/TokenContext";

import SettingsError from "@assets/settings-error.svg";
import { getCurrentURL } from "@utils/url";

export const SettingsPage: React.FC = () => {
  const { t } = useTranslation();
  const [sdk, setSDK] = useState<AppExtensionsSDK | null>();
  const { access_token: accessToken, error, status } = useSnapshot(AuthToken);
  const [admin, setAdmin] = useState(false);
  const [loading, setLoading] = useState(true);
  const [address, setAddress] = useState<string | undefined>(undefined);
  const [secret, setSecret] = useState<string | undefined>(undefined);
  const [header, setHeader] = useState<string | undefined>(undefined);
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    new AppExtensionsSDK()
      .initialize()
      .then((s) => {
        setSDK(s);
      })
      .catch(() => setSDK(null));
  }, []);

  useEffect(() => {
    if (accessToken && !error && !!sdk) {
      const { url } = getCurrentURL();
      getPipedriveMe(`${url}api/v1/users/me`)
        .then(async (ures) => {
          try {
            if (ures.data.access.find((a) => a.app === "global" && a.admin)) {
              const res = await getSettings(sdk);
              setAddress(res.doc_address);
              setSecret(res.doc_secret);
              setHeader(res.doc_header);
              setAdmin(true);
            }
          } catch {
            setAdmin(false);
          } finally {
            setLoading(false);
          }
        })
        .catch((e) => {
          setLoading(false);
        });
    }

    if (error) setLoading(false);
  }, [sdk, accessToken, error]);

  const handleSettings = async () => {
    if (address && secret && header && sdk) {
      try {
        setSaving(true);
        if (!address.endsWith("/")) {
          await postSettings(sdk, `${address}/`, secret, header);
          setAddress(`${address}/`);
        } else {
          await postSettings(sdk, address, secret, header);
        }
        await sdk.execute(Command.SHOW_SNACKBAR, {
          message: t(
            "settings.saving.ok",
            "ONLYOFFICE settings have been saved"
          ),
        });
      } catch {
        await sdk.execute(Command.SHOW_SNACKBAR, {
          message: t(
            "settings.saving.error",
            "Could not save ONLYOFFICE settings"
          ),
        });
      } finally {
        setSaving(false);
      }
    }
  };

  return (
    <div className="custom-scroll w-screen h-screen overflow-y-scroll overflow-x-hidden">
      {loading && !error && (
        <div className="h-full w-full flex justify-center items-center">
          <OnlyofficeSpinner />
        </div>
      )}
      {!loading && error && (
        <OnlyofficeBackgroundError
          Icon={<SettingsError />}
          title={t("background.error.title", "Error")}
          subtitle={t(
            status !== 401
              ? "background.error.subtitle"
              : "background.error.subtitle.token",
            status !== 401
              ? "Could not fetch plugin settings. Something went wrong. Please reload the pipedrive window"
              : "Could not fetch plugin settings. Something went wrong with your access token. Please reinstall the app"
          )}
          button={
            status === 401 && t("background.reinstall.button", "Reinstall") ||
            "Reinstall"
          }
          onClick={status === 401 ? () => {
            if (status === 401)
              window.open(
                `${getCurrentURL().url}settings/marketplace`,
                "_blank"
              );
          } : undefined}
        />
      )}
      {!loading && !error && !admin && (
        <OnlyofficeBackgroundError
          Icon={<SettingsError />}
          title={t("background.access.title", "Access Denied")}
          subtitle={t(
            "background.access.subtitle",
            "Something went wrong or access denied"
          )}
        />
      )}
      {!loading && !error && admin && (
        <>
          <div className="flex flex-col items-start pl-5 pr-5 pt-5 pb-3">
            <div className="pb-2">
              <OnlyofficeTitle
                text={t("settings.title", "Configure ONLYOFFICE app settings")}
              />
            </div>
            <p className="text-slate-800 font-normal text-base text-left">
              {t(
                "settings.text",
                `
                The plugin which enables the users to edit office documents from
                Pipedrive using ONLYOFFICE Document Server, allows multiple users
                to collaborate in real time and to save back those changes to
                Pipedrive
              `
              )}
            </p>
          </div>
          <div className="max-w-[320px]">
            <div className="pl-5 pr-5 pb-2">
              <OnlyofficeInput
                text={t("settings.inputs.address", "Document Server Address")}
                valid={!!address}
                disabled={saving}
                value={address}
                onChange={(e) => setAddress(e.target.value)}
              />
            </div>
            <div className="pl-5 pr-5 pb-2">
              <OnlyofficeInput
                text={t("settings.inputs.secret", "Document Server Secret")}
                valid={!!secret}
                disabled={saving}
                value={secret}
                onChange={(e) => setSecret(e.target.value)}
                type="password"
              />
            </div>
            <div className="pl-5 pr-5">
              <OnlyofficeInput
                text={t("settings.inputs.header", "Document Server Header")}
                valid={!!header}
                disabled={saving}
                value={header}
                onChange={(e) => setHeader(e.target.value)}
              />
            </div>
            <div className="flex justify-start items-center mt-4 ml-5">
              <OnlyofficeButton
                text={t("button.save", "Save")}
                primary
                disabled={saving}
                onClick={handleSettings}
              />
            </div>
            <div className="relative bottom-0 ml-5 w-[568px]">
              <Banner />
            </div>
          </div>
        </>
      )}
    </div>
  );
};
