export interface OpsUser {
  id: number;
  username: string;
  role: "data_ops" | "super_admin" | string;
}

export interface OpsPharmacy {
  id: number;
  slug: string;
  name: string;
  site_domain: string;
  logo_url?: string;
  status: string;
}

export interface OpsSource {
  id: number;
  pharmacy_id: number;
  pharmacy_name: string;
  pharmacy_slug: string;
  kind: string;
  config: Record<string, unknown>;
  schedule_interval: string;
  enabled: boolean;
  last_run_at?: string;
  last_status?: string;
  last_error?: string;
}

export interface PreviewItem {
  external_id: string;
  name_fa: string;
  brand_name: string;
  price_toman: number;
  in_stock: boolean;
  product_url: string;
  image_url: string;
}

export interface OpsMatch {
  id: number;
  source_item_id: number;
  source_id: number;
  external_id: string;
  score: number;
  status: string;
  reason: string;
  raw: Record<string, unknown>;
  current_product_id?: number;
  suggested_product_id?: number;
  suggested_slug?: string;
  suggested_name?: string;
  suggested_image?: string;
  created_at: string;
}

export interface OpsProduct {
  id: number;
  slug: string;
  name_fa: string;
  name_en?: string;
  generic_name?: string;
  brand_id?: number;
  brand_name?: string;
  category_id?: number;
  category_name?: string;
  image_url?: string;
  status: "published" | "hidden" | "needs_review" | string;
  locked_fields: string[];
  source_snapshot: Record<string, unknown>;
}

export interface OpsCategory {
  id: number;
  parent_id?: number;
  slug: string;
  name_fa: string;
  position: number;
}

export interface OpsBrand {
  id: number;
  slug: string;
  name_fa: string;
  name_en?: string;
}

export interface OpsAudit {
  id: number;
  actor_id?: number;
  actor_name: string;
  entity: string;
  entity_id: string;
  action: string;
  before: Record<string, unknown>;
  after: Record<string, unknown>;
  created_at: string;
}

export interface OpsJob {
  job_id?: number;
  kind?: string;
  status?: string;
  attempts?: number;
  run_at?: string;
  last_error?: string;
  enqueued: boolean;
  latest_run?: {
    id: number;
    status: string;
    started_at: string;
    finished_at?: string;
    ok_count: number;
    fail_count: number;
    error?: string;
  };
}

export interface Page<T> {
  items: T[];
  page: number;
  page_size: number;
  total: number;
}

export interface OpsHealthSource {
  id: number;
  pharmacy_id: number;
  pharmacy_name: string;
  pharmacy_slug: string;
  kind: string;
  schedule_interval: string;
  enabled: boolean;
  last_run_at?: string;
  last_status?: string;
  last_error?: string;
  overdue: boolean;
  low_trust: boolean;
  reject_count: number;
  success_rate: number;
  runs: number;
  succeeded: number;
  run_duration_ms?: number;
  latest_run?: {
    id: number;
    status: string;
    started_at: string;
    finished_at?: string;
    ok_count: number;
    fail_count: number;
    error?: string;
  };
}

export interface OpsSyncRun {
  id: number;
  status: string;
  started_at: string;
  finished_at?: string;
  ok_count: number;
  fail_count: number;
  error?: string;
}

export interface OpsSuspiciousOffer {
  id: number;
  product_id: number;
  product_slug: string;
  product_name: string;
  pharmacy_id: number;
  pharmacy_name: string;
  pharmacy_slug: string;
  source_id?: number;
  source_reject_count: number;
  current_price_rial: number;
  proposed_price_rial: number;
  proposed_at?: string;
  product_url: string;
  last_seen_at: string;
  in_stock: boolean;
}

export interface OpsStaleSource {
  id: number;
  pharmacy_name: string;
  pharmacy_slug: string;
  oldest_seen: string;
  offer_count: number;
  age_hours: number;
}

export interface OpsClickPharmacy {
  id: number;
  slug: string;
  name: string;
  clicks: number;
}

export interface OpsClickProduct {
  id: number;
  slug: string;
  name_fa: string;
  clicks: number;
}

export interface OpsClickReport {
  from: string;
  to: string;
  total: number;
  pharmacies: OpsClickPharmacy[];
  top_products: OpsClickProduct[];
  top_pharmacy?: OpsClickPharmacy | null;
}

export interface AdminDashboard {
  from: string;
  to: string;
  days: number;
  searches: number;
  clicks: number;
  ctr: number;
  coverage: {
    active_pharmacies: number;
    published_products: number;
    median_freshness_seconds: number;
  };
  pharmacies: { id: number; slug: string; name: string; clicks: number }[];
  frequent: { query: string; hits: number; last_seen: string }[];
  zero: { query: string; hits: number; last_seen: string }[];
  series: { day: string; searches: number; clicks: number }[];
}

export interface AdminUser {
  id: number;
  username: string;
  role: string;
  is_active: boolean;
  created_at: string;
}

export interface AdminFlag {
  key: string;
  pharmacy_id?: number;
  enabled: boolean;
  updated_at: string;
}

export interface AdminPharmacy {
  id: number;
  slug: string;
  name: string;
}
