export type UserResponse = {
  id: string;
  access_token: string;
  expires_at: number;
};

type PipedriveAccess = {
  app: string;
  admin: boolean;
};

export type PipedriveUserResponse = {
  success: boolean;
  data: {
    id: number;
    access: PipedriveAccess[];
    active_flag: true;
  };
};
