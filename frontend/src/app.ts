import {FetchConfig} from 'aurelia-auth';
import {Aurelia} from 'aurelia-framework';
import {Router, RouterConfiguration} from 'aurelia-router';

export class App {
  router: Router;

  configureRouter(config: RouterConfiguration, router: Router, fetch: FetchConfig) {
    config.title = 'Aurelia';
    config.map([
      { route: ['', 'welcome'], name: 'welcome',      moduleId: './welcome',      title: 'Welcome' },
      { route: 'users',         name: 'users',        moduleId: './users',        title: 'Github Users' },
      { route: 'child-router',  name: 'child-router', moduleId: './child-router', title: 'Child Router' }
    ]);

    this.router = router;
    this.fetchConfig = fetch;
  }

  activate() {
    this.fetch.configure();
  }
}
