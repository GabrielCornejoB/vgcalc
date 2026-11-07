import type { Request, Response } from "express";

import { PaginationSchema } from "@schemas";
import type { AbilityUseCase } from "@use-cases";

export class AbilityHandler {
    private readonly useCase: AbilityUseCase; 

    constructor(useCase: AbilityUseCase) {
        this.useCase = useCase;
    }

    getByPagination = async (request: Request, response: Response) => {
        const pagination = PaginationSchema.parse(request.body);

        response.status(200).json({
            data: this.useCase.getByPagination(pagination),
            pagination
        });
    }
}
