package com.example.mobile

import android.service.autofill.AutofillService
import android.service.autofill.FillRequest
import android.service.autofill.FillResponse
import android.service.autofill.SaveRequest
import android.view.autofill.AutofillId
import android.widget.RemoteViews

/**
 * AutofillService scaffolding (Phase 2 / P2).
 *
 * NOTE: This is a UI/OS-integration stub. No crypto or vault access happens here.
 * Vault reads must go through the subprocess bridge (OrcaBridge) on the Dart side
 * or a future native binding; this service only builds the autofill UI skeleton.
 *
 * Full implementation (matching vault entries to fields, saving new credentials)
 * is deferred to a later slice.
 */
class OrcaAutofillService : AutofillService() {
    override fun onFillRequest(request: FillRequest, cancellationSignal: android.os.CancellationSignal, callback: FillCallback) {
        // Stub: no datasets yet. Return an empty response so the system
        // gracefully falls back to manual entry.
        callback.onSuccess(null)
    }

    override fun onSaveRequest(request: SaveRequest, callback: SaveCallback) {
        // Stub: no-op for v1 scaffolding.
        callback.onSuccess()
    }
}
