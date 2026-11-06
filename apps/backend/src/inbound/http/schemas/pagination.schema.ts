import z from "zod";

export const PaginationSchema = z.object({
    size: z.int().positive().max(50),
    page: z.int().positive(),
})
