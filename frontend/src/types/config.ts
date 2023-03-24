type Permissions = {
  comment: boolean;
  copy: boolean;
  deleteCommentAuthorOnly: boolean;
  download: boolean;
  edit: boolean;
  editCommentAuthorOnly: boolean;
  fillForms: boolean;
  modifyContentControl: boolean;
  modifyFilter: boolean;
  print: boolean;
  review: boolean;
};

type Document = {
  fileType: string;
  key: string;
  title: string;
  url: string;
  permissions: Permissions;
};

type User = {
  id: string;
  name: string;
};

type Goback = {
  requestClost: boolean;
};

type Customization = {
  goback: Goback;
  hideRightMenu: boolean;
  plugins: boolean;
};

type EditorConfig = {
  user: User;
  callbackUrl: string;
  customization: Customization;
  lang: string;
};

export type ConfigResponse = {
  document: Document;
  documentType: string;
  editorConfig: EditorConfig;
  type: string;
  token: string;
  server_url: string;
};
