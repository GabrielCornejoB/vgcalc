import { AbilityHandler } from "#handlers";
import { AbilityUseCase } from "#use-cases";
import { AbilityRouter, HttpApp } from "#routers";
import { PrismaAbilityRepository } from "#repositories";
import { db } from "#prisma-migrate/db";

const abilityUseCase = new AbilityUseCase(new PrismaAbilityRepository(db));

const abilityHandler = new AbilityHandler(abilityUseCase);

export const httpApp = new HttpApp([
    new AbilityRouter(abilityHandler),
]);

