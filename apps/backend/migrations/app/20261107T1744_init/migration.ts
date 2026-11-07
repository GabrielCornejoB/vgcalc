#!/usr/bin/env -S node
import type { Contract as End } from '../../snapshots/2eab1de11b8ce1b65c0a23fdca13271ce3a5bbfa36c8a2f5fcb96ec67c404103/contract';
import endContract from '../../snapshots/2eab1de11b8ce1b65c0a23fdca13271ce3a5bbfa36c8a2f5fcb96ec67c404103/contract.json' with { type: 'json' };
import {
  Migration,
  MigrationCLI,
  checkExpression,
  col,
  lit,
  primaryKey,
  sql,
} from '@prisma/orm-postgres/migration';

export default class M extends Migration<never, End> {
  override readonly endContractJson = endContract;

  override get operations() {
    return [
      this.createSchema({ schema: 'public' }),
      this.createSchema({ schema: 'vgcalc' }),
      this.createTable({
        schema: 'vgcalc',
        table: 'abilities',
        columns: [
          col('display_name', 'text', { notNull: true, codecRef: { codecId: 'pg/text@1' } }),
          col('effect', 'text', { codecRef: { codecId: 'pg/text@1' } }),
          col('id', 'SERIAL', { notNull: true, codecRef: { codecId: 'pg/int4@1' } }),
          col('pokeapi_id', 'int4', { notNull: true, codecRef: { codecId: 'pg/int4@1' } }),
          col('short_effect', 'text', { codecRef: { codecId: 'pg/text@1' } }),
          col('slug', 'text', { notNull: true, codecRef: { codecId: 'pg/text@1' } }),
        ],
        constraints: [primaryKey(['id'])],
      }),
      this.createTable({
        schema: 'vgcalc',
        table: 'items',
        columns: [
          col('display_name', 'text', { notNull: true, codecRef: { codecId: 'pg/text@1' } }),
          col('effect', 'text', { codecRef: { codecId: 'pg/text@1' } }),
          col('fling_power', 'int4', {
            notNull: true,
            default: lit(0),
            codecRef: { codecId: 'pg/int4@1' },
          }),
          col('id', 'SERIAL', { notNull: true, codecRef: { codecId: 'pg/int4@1' } }),
          col('pokeapi_id', 'int4', { notNull: true, codecRef: { codecId: 'pg/int4@1' } }),
          col('short_effect', 'text', { codecRef: { codecId: 'pg/text@1' } }),
          col('slug', 'text', { notNull: true, codecRef: { codecId: 'pg/text@1' } }),
        ],
        constraints: [
          primaryKey(['id']),
          checkExpression('items_fling_power_non_negative_eca3988b', sql`fling_power >= 0`),
        ],
      }),
      this.createTable({
        schema: 'vgcalc',
        table: 'mega_evolution',
        columns: [
          col('ability_id', 'int4', { notNull: true, codecRef: { codecId: 'pg/int4@1' } }),
          col('atk', 'int4', { notNull: true, codecRef: { codecId: 'pg/int4@1' } }),
          col('base_pokemon_id', 'int4', { notNull: true, codecRef: { codecId: 'pg/int4@1' } }),
          col('def', 'int4', { notNull: true, codecRef: { codecId: 'pg/int4@1' } }),
          col('hp', 'int4', { notNull: true, codecRef: { codecId: 'pg/int4@1' } }),
          col('id', 'SERIAL', { notNull: true, codecRef: { codecId: 'pg/int4@1' } }),
          col('pokeapi_id', 'int4', { notNull: true, codecRef: { codecId: 'pg/int4@1' } }),
          col('primary_type_id', 'int4', { notNull: true, codecRef: { codecId: 'pg/int4@1' } }),
          col('required_item_id', 'int4', { codecRef: { codecId: 'pg/int4@1' } }),
          col('required_move_id', 'int4', { codecRef: { codecId: 'pg/int4@1' } }),
          col('secondary_type_id', 'int4', { codecRef: { codecId: 'pg/int4@1' } }),
          col('slug', 'text', { notNull: true, codecRef: { codecId: 'pg/text@1' } }),
          col('spa', 'int4', { notNull: true, codecRef: { codecId: 'pg/int4@1' } }),
          col('spd', 'int4', { notNull: true, codecRef: { codecId: 'pg/int4@1' } }),
          col('spe', 'int4', { notNull: true, codecRef: { codecId: 'pg/int4@1' } }),
          col('weight', 'numeric(6,2)', {
            notNull: true,
            codecRef: { codecId: 'pg/numeric@1', typeParams: { precision: 6, scale: 2 } },
          }),
        ],
        constraints: [
          primaryKey(['id']),
          checkExpression('mega_evolution_atk_positive_197b305e', sql`atk >= 1`),
          checkExpression('mega_evolution_def_positive_fd9d232a', sql`def >= 1`),
          checkExpression('mega_evolution_hp_positive_a6ecb535', sql`hp >= 1`),
          checkExpression(
            'mega_evolution_single_requirement_b28b4b5c',
            sql`(CASE WHEN required_item_id IS NOT NULL THEN 1 ELSE 0 END + CASE WHEN required_move_id IS NOT NULL THEN 1 ELSE 0 END) = 1`,
          ),
          checkExpression('mega_evolution_spa_positive_803af724', sql`spa >= 1`),
          checkExpression('mega_evolution_spd_positive_a55d07d0', sql`spd >= 1`),
          checkExpression('mega_evolution_spe_positive_b7760321', sql`spe >= 1`),
          checkExpression('mega_evolution_weight_positive_7c35b28d', sql`weight > 0`),
        ],
      }),
      this.createTable({
        schema: 'vgcalc',
        table: 'moves',
        columns: [
          col('accuracy', 'int4', { notNull: true, codecRef: { codecId: 'pg/int4@1' } }),
          col('category', 'text', { notNull: true, codecRef: { codecId: 'pg/text@1' } }),
          col('display_name', 'text', { notNull: true, codecRef: { codecId: 'pg/text@1' } }),
          col('effect', 'text', { codecRef: { codecId: 'pg/text@1' } }),
          col('id', 'SERIAL', { notNull: true, codecRef: { codecId: 'pg/int4@1' } }),
          col('is_bite', 'bool', {
            notNull: true,
            default: lit(false),
            codecRef: { codecId: 'pg/bool@1' },
          }),
          col('is_bullet', 'bool', {
            notNull: true,
            default: lit(false),
            codecRef: { codecId: 'pg/bool@1' },
          }),
          col('is_contact', 'bool', {
            notNull: true,
            default: lit(false),
            codecRef: { codecId: 'pg/bool@1' },
          }),
          col('is_pulse', 'bool', {
            notNull: true,
            default: lit(false),
            codecRef: { codecId: 'pg/bool@1' },
          }),
          col('is_punch', 'bool', {
            notNull: true,
            default: lit(false),
            codecRef: { codecId: 'pg/bool@1' },
          }),
          col('is_slicing', 'bool', {
            notNull: true,
            default: lit(false),
            codecRef: { codecId: 'pg/bool@1' },
          }),
          col('is_sound', 'bool', {
            notNull: true,
            default: lit(false),
            codecRef: { codecId: 'pg/bool@1' },
          }),
          col('is_wind', 'bool', {
            notNull: true,
            default: lit(false),
            codecRef: { codecId: 'pg/bool@1' },
          }),
          col('max_hits', 'int4', { codecRef: { codecId: 'pg/int4@1' } }),
          col('min_hits', 'int4', { codecRef: { codecId: 'pg/int4@1' } }),
          col('pokeapi_id', 'int4', { notNull: true, codecRef: { codecId: 'pg/int4@1' } }),
          col('power', 'int4', { notNull: true, codecRef: { codecId: 'pg/int4@1' } }),
          col('pp', 'int4', { notNull: true, codecRef: { codecId: 'pg/int4@1' } }),
          col('priority', 'int4', {
            notNull: true,
            default: lit(0),
            codecRef: { codecId: 'pg/int4@1' },
          }),
          col('short_effect', 'text', { codecRef: { codecId: 'pg/text@1' } }),
          col('slug', 'text', { notNull: true, codecRef: { codecId: 'pg/text@1' } }),
          col('target', 'text', { notNull: true, codecRef: { codecId: 'pg/text@1' } }),
          col('type_id', 'int4', { notNull: true, codecRef: { codecId: 'pg/int4@1' } }),
        ],
        constraints: [
          primaryKey(['id']),
          checkExpression('moves_accuracy_range_1a03c3e6', sql`accuracy BETWEEN 0 AND 100`),
          checkExpression(
            'moves_category_allowed_9de5a245',
            sql`category IN ('physical', 'special', 'status')`,
          ),
          checkExpression(
            'moves_hits_range_b6013138',
            sql`max_hits IS NULL OR min_hits IS NULL OR max_hits >= min_hits`,
          ),
          checkExpression('moves_max_hits_positive_add4ae2a', sql`max_hits >= 1`),
          checkExpression('moves_min_hits_positive_46b5efb3', sql`min_hits >= 1`),
          checkExpression('moves_power_non_negative_ee2786d9', sql`power >= 0`),
          checkExpression('moves_pp_non_negative_a480c8c4', sql`pp >= 0`),
        ],
      }),
      this.createTable({
        schema: 'vgcalc',
        table: 'natures',
        columns: [
          col('display_name', 'text', { notNull: true, codecRef: { codecId: 'pg/text@1' } }),
          col('id', 'SERIAL', { notNull: true, codecRef: { codecId: 'pg/int4@1' } }),
          col('minus_stat', 'text', { notNull: true, codecRef: { codecId: 'pg/text@1' } }),
          col('plus_stat', 'text', { notNull: true, codecRef: { codecId: 'pg/text@1' } }),
          col('pokeapi_id', 'int4', { notNull: true, codecRef: { codecId: 'pg/int4@1' } }),
          col('slug', 'text', { notNull: true, codecRef: { codecId: 'pg/text@1' } }),
        ],
        constraints: [
          primaryKey(['id']),
          checkExpression(
            'natures_minus_stat_allowed_c5cd20cf',
            sql`minus_stat IN ('atk', 'def', 'spa', 'spd', 'spe')`,
          ),
          checkExpression(
            'natures_plus_stat_allowed_54b86585',
            sql`plus_stat IN ('atk', 'def', 'spa', 'spd', 'spe')`,
          ),
        ],
      }),
      this.createTable({
        schema: 'vgcalc',
        table: 'pokemon',
        columns: [
          col('ability1_id', 'int4', { notNull: true, codecRef: { codecId: 'pg/int4@1' } }),
          col('ability2_id', 'int4', { codecRef: { codecId: 'pg/int4@1' } }),
          col('atk', 'int4', { notNull: true, codecRef: { codecId: 'pg/int4@1' } }),
          col('def', 'int4', { notNull: true, codecRef: { codecId: 'pg/int4@1' } }),
          col('display_name', 'text', { notNull: true, codecRef: { codecId: 'pg/text@1' } }),
          col('fixed_gender', 'text', { codecRef: { codecId: 'pg/text@1' } }),
          col('hidden_ability_id', 'int4', { codecRef: { codecId: 'pg/int4@1' } }),
          col('hp', 'int4', { notNull: true, codecRef: { codecId: 'pg/int4@1' } }),
          col('id', 'SERIAL', { notNull: true, codecRef: { codecId: 'pg/int4@1' } }),
          col('pokeapi_id', 'int4', { notNull: true, codecRef: { codecId: 'pg/int4@1' } }),
          col('primary_type_id', 'int4', { notNull: true, codecRef: { codecId: 'pg/int4@1' } }),
          col('secondary_type_id', 'int4', { codecRef: { codecId: 'pg/int4@1' } }),
          col('slug', 'text', { notNull: true, codecRef: { codecId: 'pg/text@1' } }),
          col('spa', 'int4', { notNull: true, codecRef: { codecId: 'pg/int4@1' } }),
          col('spd', 'int4', { notNull: true, codecRef: { codecId: 'pg/int4@1' } }),
          col('spe', 'int4', { notNull: true, codecRef: { codecId: 'pg/int4@1' } }),
          col('weight', 'numeric(6,2)', {
            notNull: true,
            codecRef: { codecId: 'pg/numeric@1', typeParams: { precision: 6, scale: 2 } },
          }),
        ],
        constraints: [
          primaryKey(['id']),
          checkExpression('pokemon_atk_positive_197b305e', sql`atk >= 1`),
          checkExpression('pokemon_def_positive_fd9d232a', sql`def >= 1`),
          checkExpression(
            'pokemon_fixed_gender_allowed_782c5d90',
            sql`fixed_gender IN ('male', 'female', 'genderless')`,
          ),
          checkExpression('pokemon_hp_positive_a6ecb535', sql`hp >= 1`),
          checkExpression('pokemon_spa_positive_803af724', sql`spa >= 1`),
          checkExpression('pokemon_spd_positive_a55d07d0', sql`spd >= 1`),
          checkExpression('pokemon_spe_positive_b7760321', sql`spe >= 1`),
          checkExpression('pokemon_weight_positive_7c35b28d', sql`weight > 0`),
        ],
      }),
      this.createTable({
        schema: 'vgcalc',
        table: 'pokemon_moves',
        columns: [
          col('move_id', 'int4', { notNull: true, codecRef: { codecId: 'pg/int4@1' } }),
          col('pokemon_id', 'int4', { notNull: true, codecRef: { codecId: 'pg/int4@1' } }),
        ],
        constraints: [primaryKey(['pokemon_id', 'move_id'])],
      }),
      this.createTable({
        schema: 'vgcalc',
        table: 'regulation_pokemon',
        columns: [
          col('pokemon_id', 'int4', { notNull: true, codecRef: { codecId: 'pg/int4@1' } }),
          col('regulation_id', 'int4', { notNull: true, codecRef: { codecId: 'pg/int4@1' } }),
        ],
        constraints: [primaryKey(['regulation_id', 'pokemon_id'])],
      }),
      this.createTable({
        schema: 'vgcalc',
        table: 'regulations',
        columns: [
          col('display_name', 'text', { notNull: true, codecRef: { codecId: 'pg/text@1' } }),
          col('id', 'SERIAL', { notNull: true, codecRef: { codecId: 'pg/int4@1' } }),
          col('is_current', 'bool', {
            notNull: true,
            default: lit(false),
            codecRef: { codecId: 'pg/bool@1' },
          }),
          col('slug', 'text', { notNull: true, codecRef: { codecId: 'pg/text@1' } }),
        ],
        constraints: [primaryKey(['id'])],
      }),
      this.createTable({
        schema: 'vgcalc',
        table: 'type_efficacy',
        columns: [
          col('attack_type_id', 'int4', { notNull: true, codecRef: { codecId: 'pg/int4@1' } }),
          col('multiplier', 'numeric(3,1)', {
            notNull: true,
            codecRef: { codecId: 'pg/numeric@1', typeParams: { precision: 3, scale: 1 } },
          }),
          col('target_type_id', 'int4', { notNull: true, codecRef: { codecId: 'pg/int4@1' } }),
        ],
        constraints: [
          primaryKey(['attack_type_id', 'target_type_id']),
          checkExpression(
            'type_efficacy_multiplier_allowed_b4b6ef93',
            sql`multiplier IN (0, 0.5, 1, 2)`,
          ),
        ],
      }),
      this.createTable({
        schema: 'vgcalc',
        table: 'types',
        columns: [
          col('display_name', 'text', { notNull: true, codecRef: { codecId: 'pg/text@1' } }),
          col('id', 'SERIAL', { notNull: true, codecRef: { codecId: 'pg/int4@1' } }),
          col('pokeapi_id', 'int4', { notNull: true, codecRef: { codecId: 'pg/int4@1' } }),
          col('slug', 'text', { notNull: true, codecRef: { codecId: 'pg/text@1' } }),
        ],
        constraints: [primaryKey(['id'])],
      }),
      this.addUnique({
        schema: 'vgcalc',
        table: 'abilities',
        constraint: 'abilities_pokeapi_id_key',
        columns: ['pokeapi_id'],
      }),
      this.addUnique({
        schema: 'vgcalc',
        table: 'abilities',
        constraint: 'abilities_slug_key',
        columns: ['slug'],
      }),
      this.addUnique({
        schema: 'vgcalc',
        table: 'items',
        constraint: 'items_pokeapi_id_key',
        columns: ['pokeapi_id'],
      }),
      this.addUnique({
        schema: 'vgcalc',
        table: 'items',
        constraint: 'items_slug_key',
        columns: ['slug'],
      }),
      this.addUnique({
        schema: 'vgcalc',
        table: 'mega_evolution',
        constraint: 'mega_evolution_pokeapi_id_key',
        columns: ['pokeapi_id'],
      }),
      this.addUnique({
        schema: 'vgcalc',
        table: 'mega_evolution',
        constraint: 'mega_evolution_slug_key',
        columns: ['slug'],
      }),
      this.addUnique({
        schema: 'vgcalc',
        table: 'moves',
        constraint: 'moves_pokeapi_id_key',
        columns: ['pokeapi_id'],
      }),
      this.addUnique({
        schema: 'vgcalc',
        table: 'moves',
        constraint: 'moves_slug_key',
        columns: ['slug'],
      }),
      this.addUnique({
        schema: 'vgcalc',
        table: 'natures',
        constraint: 'natures_pokeapi_id_key',
        columns: ['pokeapi_id'],
      }),
      this.addUnique({
        schema: 'vgcalc',
        table: 'natures',
        constraint: 'natures_slug_key',
        columns: ['slug'],
      }),
      this.addUnique({
        schema: 'vgcalc',
        table: 'pokemon',
        constraint: 'pokemon_pokeapi_id_key',
        columns: ['pokeapi_id'],
      }),
      this.addUnique({
        schema: 'vgcalc',
        table: 'pokemon',
        constraint: 'pokemon_slug_key',
        columns: ['slug'],
      }),
      this.addUnique({
        schema: 'vgcalc',
        table: 'regulations',
        constraint: 'regulations_slug_key',
        columns: ['slug'],
      }),
      this.addUnique({
        schema: 'vgcalc',
        table: 'types',
        constraint: 'types_pokeapi_id_key',
        columns: ['pokeapi_id'],
      }),
      this.addUnique({
        schema: 'vgcalc',
        table: 'types',
        constraint: 'types_slug_key',
        columns: ['slug'],
      }),
      this.createIndex({
        schema: 'vgcalc',
        table: 'mega_evolution',
        index: 'idx_mega_base',
        columns: ['base_pokemon_id'],
      }),
      this.createIndex({
        schema: 'vgcalc',
        table: 'mega_evolution',
        index: 'mega_evolution_ability_id_idx_a0e402a2',
        columns: ['ability_id'],
      }),
      this.createIndex({
        schema: 'vgcalc',
        table: 'mega_evolution',
        index: 'mega_evolution_primary_type_id_idx_73cc02d2',
        columns: ['primary_type_id'],
      }),
      this.createIndex({
        schema: 'vgcalc',
        table: 'mega_evolution',
        index: 'mega_evolution_required_item_id_idx_fa4e92d8',
        columns: ['required_item_id'],
      }),
      this.createIndex({
        schema: 'vgcalc',
        table: 'mega_evolution',
        index: 'mega_evolution_required_move_id_idx_3d623c87',
        columns: ['required_move_id'],
      }),
      this.createIndex({
        schema: 'vgcalc',
        table: 'mega_evolution',
        index: 'mega_evolution_secondary_type_id_idx_4300e104',
        columns: ['secondary_type_id'],
      }),
      this.createIndex({
        schema: 'vgcalc',
        table: 'moves',
        index: 'idx_moves_type',
        columns: ['type_id'],
      }),
      this.createIndex({
        schema: 'vgcalc',
        table: 'pokemon',
        index: 'idx_pokemon_ability1',
        columns: ['ability1_id'],
      }),
      this.createIndex({
        schema: 'vgcalc',
        table: 'pokemon',
        index: 'idx_pokemon_primary_type',
        columns: ['primary_type_id'],
      }),
      this.createIndex({
        schema: 'vgcalc',
        table: 'pokemon',
        index: 'pokemon_ability2_id_idx_76e8bff6',
        columns: ['ability2_id'],
      }),
      this.createIndex({
        schema: 'vgcalc',
        table: 'pokemon',
        index: 'pokemon_hidden_ability_id_idx_cda8a618',
        columns: ['hidden_ability_id'],
      }),
      this.createIndex({
        schema: 'vgcalc',
        table: 'pokemon',
        index: 'pokemon_secondary_type_id_idx_4300e104',
        columns: ['secondary_type_id'],
      }),
      this.createIndex({
        schema: 'vgcalc',
        table: 'pokemon_moves',
        index: 'idx_pokemon_moves_move',
        columns: ['move_id'],
      }),
      this.createIndex({
        schema: 'vgcalc',
        table: 'pokemon_moves',
        index: 'pokemon_moves_pokemon_id_idx_c2281696',
        columns: ['pokemon_id'],
      }),
      this.createIndex({
        schema: 'vgcalc',
        table: 'regulation_pokemon',
        index: 'idx_regulation_pokemon_pokemon',
        columns: ['pokemon_id'],
      }),
      this.createIndex({
        schema: 'vgcalc',
        table: 'regulation_pokemon',
        index: 'regulation_pokemon_regulation_id_idx_7cdab16a',
        columns: ['regulation_id'],
      }),
      this.createIndex({
        schema: 'vgcalc',
        table: 'regulations',
        index: 'regulations_single_current_e77f80a7',
        columns: ['is_current'],
        extras: { where: sql`is_current = TRUE`, unique: true },
      }),
      this.createIndex({
        schema: 'vgcalc',
        table: 'type_efficacy',
        index: 'type_efficacy_attack_type_id_idx_47ca91d3',
        columns: ['attack_type_id'],
      }),
      this.createIndex({
        schema: 'vgcalc',
        table: 'type_efficacy',
        index: 'type_efficacy_target_type_id_idx_f421982a',
        columns: ['target_type_id'],
      }),
      this.addForeignKey({
        schema: 'vgcalc',
        table: 'mega_evolution',
        foreignKey: {
          name: 'mega_evolution_base_pokemon_id_fkey',
          columns: ['base_pokemon_id'],
          references: { schema: 'vgcalc', table: 'pokemon', columns: ['id'] },
        },
      }),
      this.addForeignKey({
        schema: 'vgcalc',
        table: 'mega_evolution',
        foreignKey: {
          name: 'mega_evolution_required_item_id_fkey',
          columns: ['required_item_id'],
          references: { schema: 'vgcalc', table: 'items', columns: ['id'] },
        },
      }),
      this.addForeignKey({
        schema: 'vgcalc',
        table: 'mega_evolution',
        foreignKey: {
          name: 'mega_evolution_required_move_id_fkey',
          columns: ['required_move_id'],
          references: { schema: 'vgcalc', table: 'moves', columns: ['id'] },
        },
      }),
      this.addForeignKey({
        schema: 'vgcalc',
        table: 'mega_evolution',
        foreignKey: {
          name: 'mega_evolution_ability_id_fkey',
          columns: ['ability_id'],
          references: { schema: 'vgcalc', table: 'abilities', columns: ['id'] },
        },
      }),
      this.addForeignKey({
        schema: 'vgcalc',
        table: 'mega_evolution',
        foreignKey: {
          name: 'mega_evolution_primary_type_id_fkey',
          columns: ['primary_type_id'],
          references: { schema: 'vgcalc', table: 'types', columns: ['id'] },
        },
      }),
      this.addForeignKey({
        schema: 'vgcalc',
        table: 'mega_evolution',
        foreignKey: {
          name: 'mega_evolution_secondary_type_id_fkey',
          columns: ['secondary_type_id'],
          references: { schema: 'vgcalc', table: 'types', columns: ['id'] },
        },
      }),
      this.addForeignKey({
        schema: 'vgcalc',
        table: 'moves',
        foreignKey: {
          name: 'moves_type_id_fkey',
          columns: ['type_id'],
          references: { schema: 'vgcalc', table: 'types', columns: ['id'] },
        },
      }),
      this.addForeignKey({
        schema: 'vgcalc',
        table: 'pokemon',
        foreignKey: {
          name: 'pokemon_primary_type_id_fkey',
          columns: ['primary_type_id'],
          references: { schema: 'vgcalc', table: 'types', columns: ['id'] },
        },
      }),
      this.addForeignKey({
        schema: 'vgcalc',
        table: 'pokemon',
        foreignKey: {
          name: 'pokemon_secondary_type_id_fkey',
          columns: ['secondary_type_id'],
          references: { schema: 'vgcalc', table: 'types', columns: ['id'] },
        },
      }),
      this.addForeignKey({
        schema: 'vgcalc',
        table: 'pokemon',
        foreignKey: {
          name: 'pokemon_ability1_id_fkey',
          columns: ['ability1_id'],
          references: { schema: 'vgcalc', table: 'abilities', columns: ['id'] },
        },
      }),
      this.addForeignKey({
        schema: 'vgcalc',
        table: 'pokemon',
        foreignKey: {
          name: 'pokemon_ability2_id_fkey',
          columns: ['ability2_id'],
          references: { schema: 'vgcalc', table: 'abilities', columns: ['id'] },
        },
      }),
      this.addForeignKey({
        schema: 'vgcalc',
        table: 'pokemon',
        foreignKey: {
          name: 'pokemon_hidden_ability_id_fkey',
          columns: ['hidden_ability_id'],
          references: { schema: 'vgcalc', table: 'abilities', columns: ['id'] },
        },
      }),
      this.addForeignKey({
        schema: 'vgcalc',
        table: 'pokemon_moves',
        foreignKey: {
          name: 'pokemon_moves_pokemon_id_fkey',
          columns: ['pokemon_id'],
          references: { schema: 'vgcalc', table: 'pokemon', columns: ['id'] },
          onDelete: 'cascade',
        },
      }),
      this.addForeignKey({
        schema: 'vgcalc',
        table: 'pokemon_moves',
        foreignKey: {
          name: 'pokemon_moves_move_id_fkey',
          columns: ['move_id'],
          references: { schema: 'vgcalc', table: 'moves', columns: ['id'] },
          onDelete: 'cascade',
        },
      }),
      this.addForeignKey({
        schema: 'vgcalc',
        table: 'regulation_pokemon',
        foreignKey: {
          name: 'regulation_pokemon_regulation_id_fkey',
          columns: ['regulation_id'],
          references: { schema: 'vgcalc', table: 'regulations', columns: ['id'] },
          onDelete: 'cascade',
        },
      }),
      this.addForeignKey({
        schema: 'vgcalc',
        table: 'regulation_pokemon',
        foreignKey: {
          name: 'regulation_pokemon_pokemon_id_fkey',
          columns: ['pokemon_id'],
          references: { schema: 'vgcalc', table: 'pokemon', columns: ['id'] },
          onDelete: 'cascade',
        },
      }),
      this.addForeignKey({
        schema: 'vgcalc',
        table: 'type_efficacy',
        foreignKey: {
          name: 'type_efficacy_attack_type_id_fkey',
          columns: ['attack_type_id'],
          references: { schema: 'vgcalc', table: 'types', columns: ['id'] },
        },
      }),
      this.addForeignKey({
        schema: 'vgcalc',
        table: 'type_efficacy',
        foreignKey: {
          name: 'type_efficacy_target_type_id_fkey',
          columns: ['target_type_id'],
          references: { schema: 'vgcalc', table: 'types', columns: ['id'] },
        },
      }),
    ];
  }
}

MigrationCLI.run(import.meta.url, M);
