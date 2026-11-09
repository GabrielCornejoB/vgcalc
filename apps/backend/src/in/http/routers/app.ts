import express, { type Express, type Router } from 'express';

export interface HttpRouter {
    readonly basePath: string;
    readonly router: Router;
}

export class HttpApp {
    readonly express: Express = express();

    constructor(routers: HttpRouter[]) {
        this.express.use(express.json());
        routers.forEach((router) => this.express.use(router.basePath, router.router));
    }
}
