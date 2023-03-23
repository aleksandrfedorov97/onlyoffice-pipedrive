export type File = {
  id: string;
  user_id: number;
  name: string;
  file_size: number;
  file_type: string;
  add_time: string;
  update_time: string;
  url: string;
  person_name: string;
  remote_location: string;
};

type Pagination = {
  pagination: {
    start: number;
    next_start: number;
    limit: number;
    more_items_in_collection: boolean;
  };
};

export type FileResponse = {
  success: boolean;
  data: File[];
  additional_data: Pagination;
};
