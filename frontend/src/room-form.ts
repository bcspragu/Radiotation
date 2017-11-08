import {autoinject} from 'aurelia-framework';
import {HttpClient, json} from 'aurelia-fetch-client';

export interface Sourcer {
   id: string;
   displayName: string;
}

export interface Shuffler {
   id: string;
   displayName: string;
}

@autoinject
export class RoomForm {
  sourcers: Sourcer[] = [
    { id: 'spotify', displayName: 'Spotify' },
  ];

  shufflers: Shuffler[] = [
    { id: 'robin', displayName: 'Round Robin' },
    { id: 'shuffle', displayName: 'Fair Random' },
    { id: 'random', displayName: 'True Random' },
  ];

  musicMatcher = (a, b) => a.id === b.id;

  roomName: string = '';
  musicSource: Sourcer = this.sourcers[0];
  shuffleOrder: Shuffler = this.shufflers[0];

  constructor(private http: HttpClient) {}

  createRoom() {
    console.log(this.roomName, this.musicSource, this.shuffleOrder);
    this.http.fetch('rooms', {
      method: 'post',
      body: json({
        'music_source': this.musicSource.id,
        'shuffle_order': this.shuffleOrder.id,
        'room': this.roomName,
      })
    })
  }
}
