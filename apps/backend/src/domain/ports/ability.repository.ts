import type { Ability, Pagination } from '@models';

export interface AbilityRepository {
    findByPagination(pagination: Pagination): Promise<Ability[]>;
}
