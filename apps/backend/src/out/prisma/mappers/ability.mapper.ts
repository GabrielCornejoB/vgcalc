import type { Ability } from "#models";
import type { AbilityEntity } from "#entities";

export const toDomainAbility = (entity: AbilityEntity): Ability => ({
    pokeapiId: entity.pokeapiId,
    slug: entity.slug,
    displayName: entity.displayName,
    shortEffect: entity.shortEffect || "",
    effect: entity.effect || "",
});
