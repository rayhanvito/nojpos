import { describe, expect, it } from 'vitest';

import type { ApiEnvelope } from '../api/envelope';
import { mapSettingsEnvelope, mapSubscriptionEnvelope } from './settings-readonly';
import { createSessionPayload } from './session-cookie';

const session = createSessionPayload({
  token: 'private-session-value',
  user: { id: 'owner-1', name: 'Owner', role: 'owner', business_id: 'business-1', is_superadmin: false },
  business: { id: 'business-1', name: 'Kopi Senja' },
  device_uuid: 'web-device-1',
});

describe('settings/subscription read-only mapper', () => {
  it('maps settings fields and omits private field names', () => {
    const envelope: ApiEnvelope<Record<string, unknown>> = {
      data: { name: 'Kopi Senja', timezone: 'Asia/Jakarta', private_marker: 'hidden-value' },
      meta: { request_id: 'req-settings-unit' },
    };

    const mapped = mapSettingsEnvelope(envelope, session, 'business');
    const serialized = JSON.stringify(mapped);

    expect(mapped.data.can_edit).toBe(false);
    expect(mapped.data.fields.map((field) => field.label)).toContain('name');
    expect(serialized).not.toContain('private-session-value');
  });

  it('maps subscription without exposing session material', () => {
    const envelope: ApiEnvelope<Record<string, unknown>> = {
      data: { plan_name: 'Pro', status: 'active', can_change_plan: true },
      meta: { request_id: 'req-subscription-unit' },
    };

    const mapped = mapSubscriptionEnvelope(envelope, session);
    const serialized = JSON.stringify(mapped);

    expect(mapped.data.meta.business_id).toBe('business-1');
    expect(mapped.data.data_notes[0].message).toContain('disabled');
    expect(serialized).not.toContain('private-session-value');
    expect(serialized.toLowerCase()).not.toContain('bearer');
  });
});
