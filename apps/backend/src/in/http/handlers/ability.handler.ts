import type { Request, Response } from 'express';

import { PaginationSchema } from '#schemas';
import type { AbilityUseCase } from '#use-cases';

export class AbilityHandler {
    private readonly useCase: AbilityUseCase;

    constructor(useCase: AbilityUseCase) {
        this.useCase = useCase;
    }

    getByPagination = async (request: Request, response: Response) => {
        // TODO: invalid input throws a ZodError that Express answers as a 500 HTML page; return a 400 JSON instead (see specs/backend-pending-changes.md §1)
        const pagination = PaginationSchema.parse(request.body);

        // TODO: include the total count so clients know how many pages exist (see specs/backend-pending-changes.md §2)
        response.status(200).json({
            data: await this.useCase.getByPagination(pagination),
            pagination,
        });
    };
}
