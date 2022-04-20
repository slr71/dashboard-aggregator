FROM node:16-alpine

COPY . /src
WORKDIR /src

RUN npm ci
RUN npm run build

EXPOSE 3000

CMD ["npm", "start"]

LABEL org.label-schema.vcs-url="https://github.com/cyverse-de/dashboard-aggregator"
