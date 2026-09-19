export interface SiteSEO {
  title: string;
  description: string;
}

export interface SiteBanner {
  enabled: boolean;
  title: string;
  body: string;
}

export interface SitePages {
  about: string;
  contact: string;
  terms: string;
  disclaimer: string;
}

export interface SiteSettings {
  seo: SiteSEO;
  banner: SiteBanner;
  featured_categories: { slugs: string[] };
  pages: SitePages;
}

export const fallbackSite: SiteSettings = {
  seo: {
    title: "مقایسه قیمت دارو",
    description:
      "قیمت دارو و محصولات سلامت را در چند داروخانه آنلاین مقایسه کنید و برای خرید به سایت همان داروخانه بروید.",
  },
  banner: { enabled: false, title: "", body: "" },
  featured_categories: { slugs: [] },
  pages: {
    about:
      "این وب‌سایت قیمت دارو و محصولات سلامت را از داروخانه‌های آنلاین جمع می‌کند تا بتوانید مقایسه کنید و برای خرید به سایت همان داروخانه بروید. ما فروشنده نیستیم و سفارشی ثبت نمی‌کنیم.",
    contact:
      "برای گزارش خطای داده یا درخواست منبع جدید، از طریق ایمیل عملیاتی اعلام‌شده در استقرار پیام بفرستید. این نشانی برای مشاوره درمانی نیست.",
    terms:
      "استفاده از این وب‌سایت به معنای پذیرش این قواعد است: اطلاعات قیمت از منابع داروخانه‌ها می‌آید و ممکن است تا چند ساعت کهنه باشد. خرید و پرداخت فقط در سایت داروخانه انجام می‌شود. پلتفرم واسطه فروش، سبد خرید یا پرداخت نیست.",
    disclaimer:
      "این وب‌سایت مرجع تشخیص یا درمان نیست و توصیه پزشکی یا دارویی ارائه نمی‌دهد. تصمیم درمانی را با پزشک یا داروساز بگیرید. قیمت و موجودی نهایی فقط در سایت داروخانه معتبر است.",
  },
};
