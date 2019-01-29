export interface Room {
  id: string;
  displayName: string;
}

export interface Artist {
  name: string;
}

export interface Album {
  name: string;
  images: Array<{url: string}>;
}

export interface Track {
  name: string;
  artists: Artist[];
  album: Album;
}

export interface QueueTrack {
  id: string;
  played: boolean;
  track: Track;
}
