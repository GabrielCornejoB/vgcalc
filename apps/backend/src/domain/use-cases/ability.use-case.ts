import type { Ability, Pagination } from '#models'
import { type AbilityRepository } from '#ports'

export class AbilityUseCase {
    private readonly repository: AbilityRepository;

    constructor(repository: AbilityRepository) {
        this.repository = repository;
    }

    async getByPagination(pagination: Pagination): Promise<Ability[]> {
        const abilities = await this.repository.findByPagination(pagination);
        return abilities;
    }
}
