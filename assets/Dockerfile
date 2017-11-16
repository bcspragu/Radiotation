FROM node:alpine

EXPOSE 8081

RUN npm install --global npx
RUN mkdir -p /project

WORKDIR /project
RUN npm init --yes && \
    npm install --save \
      vue \
      vue-router \
      spectre.css \
      jquery

RUN npm install --save-dev \
      webpack \
      webpack-dev-server \
      uglifyjs-webpack-plugin \
      css-loader \
      file-loader \
      vue-cli \
      vue-loader \
      vue-template-compiler \
      extract-text-webpack-plugin

VOLUME /project/assets

CMD ["npx", "webpack-dev-server", "--config", "assets/webpack.config.js"]
