import { Router } from "express";
import type { AbilityHandler } from "@handlers";
import type { HttpRouter } from "./app.ts";

export class AbilityRouter implements HttpRouter {
    readonly basePath = "/abilities";
    readonly router = Router();

    constructor(handler: AbilityHandler) {
        if (!this.router.query) return;
        this.router.query("/", handler.getByPagination);
    }
}
