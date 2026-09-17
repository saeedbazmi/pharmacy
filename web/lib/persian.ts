/**
 * Persian text normalisation applied to user input before it is sent to the
 * backend. The backend normalises again with the same rules: the frontend copy
 * is for a better search experience, the backend copy is the authority.
 */

const PERSIAN_DIGITS = "۰۱۲۳۴۵۶۷۸۹";
const ARABIC_DIGITS = "٠١٢٣٤٥٦٧٨٩";

/** Converts Persian and Arabic-Indic digits to their Latin equivalents. */
export function toLatinDigits(input: string): string {
  return input.replace(/[۰-۹٠-٩]/g, (digit) => {
    const persian = PERSIAN_DIGITS.indexOf(digit);
    if (persian !== -1) return String(persian);
    return String(ARABIC_DIGITS.indexOf(digit));
  });
}

/**
 * Normalises Persian text so that different spellings of the same word match:
 * Arabic ye/kaf become Persian, zero-width characters are dropped, the
 * zero-width non-joiner becomes a plain space and digits become Latin.
 */
export function normalizePersian(input: string): string {
  return toLatinDigits(input)
    .replace(/[\u064A\u0649]/g, "\u06CC") // Arabic ye -> Persian ye
    .replace(/\u0643/g, "\u06A9") // Arabic kaf -> Persian kaf
    .replace(/[\u0623\u0625\u0622]/g, "\u0627") // hamza forms -> alef
    .replace(/\u0629/g, "\u0647") // te marbuta -> he
    .replace(/[\u064B-\u0652\u0670]/g, "") // harakat
    .replace(/\u200C/g, " ") // ZWNJ -> space
    .replace(/[\u200B\u200D\u200E\u200F\uFEFF]/g, "") // other zero-width marks
    .replace(/\s+/g, " ")
    .trim();
}

/** Prepares a search query: normalised, and capped to a sane length. */
export function normalizeQuery(input: string, maxLength = 120): string {
  return normalizePersian(input).slice(0, maxLength);
}
