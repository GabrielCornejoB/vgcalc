import { AbilityHandler } from '#handlers';
import { db } from '#prisma-migrate/db';
import { PrismaAbilityRepository } from '#repositories';
import { AbilityRouter, HttpApp } from '#routers';
import { AbilityUseCase } from '#use-cases';

const abilityUseCase = new AbilityUseCase(new PrismaAbilityRepository(db));

const abilityHandler = new AbilityHandler(abilityUseCase);

export const httpApp = new HttpApp([new AbilityRouter(abilityHandler)]);
