export const getCurrentURL = () => {
  const url =
    window.location !== window.parent.location
      ? document.referrer
      : document.location.href;
  const params = new URL(document.location.href).searchParams;

  return { url, parameters: params };
};
