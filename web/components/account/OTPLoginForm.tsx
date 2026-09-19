"use client";

import { useActionState, useEffect, useState } from "react";

import { requestOTP, verifyOTP, type OTPState } from "@/app/(public)/account/actions";
import { SubmitButton } from "@/components/account/SubmitButton";

export function OTPLoginForm({
  nextPath,
  intent,
  productId,
}: {
  nextPath: string;
  intent: string;
  productId: string;
}) {
  const [requestState, requestAction] = useActionState(requestOTP, {} as OTPState);
  const [verifyState, verifyAction] = useActionState(verifyOTP, {} as OTPState);
  const sent = requestState.sent === true;
  const phone = requestState.phone ?? "";
  const error = verifyState.error || requestState.error;
  const [seconds, setSeconds] = useState(0);

  useEffect(() => {
    if (!sent) return;
    setSeconds(requestState.resendAfter ?? 60);
  }, [sent, requestState.resendAfter, requestState.phoneMasked]);

  useEffect(() => {
    if (seconds <= 0) return;
    const id = window.setInterval(() => {
      setSeconds((n) => (n > 0 ? n - 1 : 0));
    }, 1000);
    return () => window.clearInterval(id);
  }, [seconds]);

  return (
    <div className="flex flex-col gap-4 rounded-xl border border-border bg-surface p-6">
      {!sent ? (
        <form action={requestAction} className="flex flex-col gap-4">
          <label className="flex flex-col gap-1 text-sm" htmlFor="phone">
            شماره موبایل
            <input
              id="phone"
              name="phone"
              type="tel"
              inputMode="numeric"
              autoComplete="tel"
              required
              className="rounded-lg border border-border px-3 py-2"
              placeholder="۰۹۱۲۱۲۳۴۵۶۷"
            />
          </label>
          {error ? (
            <p className="text-sm text-danger" role="alert">
              {error}
            </p>
          ) : null}
          <SubmitButton pendingLabel="در حال ارسال…">ارسال کد</SubmitButton>
        </form>
      ) : (
        <form action={verifyAction} className="flex flex-col gap-4">
          <p className="text-sm text-muted">
            کد به {requestState.phoneMasked ?? "شماره شما"} ارسال شد.
          </p>
          <input type="hidden" name="phone" value={phone} />
          <input type="hidden" name="next" value={nextPath} />
          <input type="hidden" name="intent" value={intent} />
          <input type="hidden" name="product_id" value={productId} />
          <label className="flex flex-col gap-1 text-sm" htmlFor="code">
            کد تایید
            <input
              id="code"
              name="code"
              type="tel"
              inputMode="numeric"
              autoComplete="one-time-code"
              required
              maxLength={6}
              className="rounded-lg border border-border px-3 py-2 tracking-widest"
            />
          </label>
          {error ? (
            <p className="text-sm text-danger" role="alert">
              {error}
            </p>
          ) : null}
          <SubmitButton pendingLabel="در حال بررسی…">ورود</SubmitButton>
        </form>
      )}

      {sent ? (
        <form action={requestAction}>
          <input type="hidden" name="phone" value={phone} />
          <button
            type="submit"
            disabled={seconds > 0}
            className="text-sm text-primary hover:text-primary-dark disabled:cursor-not-allowed disabled:text-muted"
          >
            {seconds > 0 ? `ارسال دوباره تا ${seconds} ثانیه دیگر` : "ارسال دوباره کد"}
          </button>
        </form>
      ) : null}
    </div>
  );
}
