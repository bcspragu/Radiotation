import {Aurelia} from 'aurelia-framework';
import {HttpClient} from 'aurelia-fetch-client';
import {Config} from './config.ts'
import '../styles/styles.css';

// comment out if you don't want a Promise polyfill (remove also from webpack.config.js)
import * as Bluebird from 'bluebird';
Bluebird.config({ warnings: false });

export async function configure(aurelia: Aurelia) {
  aurelia.use
    .standardConfiguration()
    .developmentLogging()
    .plugin('aurelia-auth', (baseConfig) => {
      baseConfig.configure(Config)
    });

  // Uncomment the line below to enable animation.
  // aurelia.use.plugin('aurelia-animator-css');
  // if the css animator is enabled, add swap-order="after" to all router-view elements

  // Anyone wanting to use HTMLImports to load views, will need to install the following plugin.
  // aurelia.use.plugin('aurelia-html-import-template-loader')

  configureContainer(aurelia.container);
  await aurelia.start();
  aurelia.setRoot('app');

  // if you would like your website to work offline (Service Worker), 
  // install and enable the @easy-webpack/config-offline package in webpack.config.js and uncomment the following code:
  /*
  const offline = await System.import('offline-plugin/runtime');
  offline.install();
  */
}

function configureContainer(container) {
  let http = new HttpClient();
  http.configure(config => {
    config
      .useStandardConfiguration()
      .withInterceptor({
        request(request) {
          console.log(`Requesting ${request.method} ${request.url}`);
          return request; // you can return a modified Request, or you can short-circuit the request by returning a Response
        },
        response(response) {
          console.log(`Received ${response.status} ${response.url}`);
          return response; // you can return a modified Response
        }
      });
  });
  container.registerInstance(HttpClient, http);
}
