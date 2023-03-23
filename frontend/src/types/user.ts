export type UserResponse = {
  id: string;
  access_token: string;
  expires_at: number;
};

type PipedriveAccess = {
  app: string;
  admin: boolean;
};

type PipedriveUser = {
  id: number;
  name: string;
}

export type PipedriveUserResponse = {
  success: boolean;
  data: {
    id: number;
    name: string;
    access: PipedriveAccess[];
    active_flag: true;
  };
};

export type PipedriveSearchUsersResponse = {
  success: boolean;
  data: PipedriveUser[];
};
