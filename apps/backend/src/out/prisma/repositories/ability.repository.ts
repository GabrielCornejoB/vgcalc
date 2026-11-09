import type { Pagination } from '#models';
import { toDomainAbility } from '#out-mappers';
import type { AbilityRepository } from '#ports';
import type { db } from '#prisma-migrate/db';

type Database = typeof db;

export class PrismaAbilityRepository implements AbilityRepository {
    private readonly db: Database;

    constructor(db: Database) {
        this.db = db;
    }

    async findByPagination({ page, size }: Pagination) {
        const rows = await this.db.orm.vgcalc.Ability.orderBy((r) => r.id.asc())
            .offset(page * size)
            .limit(size)
            .all();

        return rows.map(toDomainAbility);
    }
}
