import { AbilityHandler } from "@handlers";
import { AbilityUseCase } from "@use-cases";
import { AbilityRouter, HttpApp } from "@routers";

const abilityUseCase = new AbilityUseCase({} as any);

const abilityHandler = new AbilityHandler(abilityUseCase);

export const httpApp = new HttpApp([
    new AbilityRouter(abilityHandler),
]);

