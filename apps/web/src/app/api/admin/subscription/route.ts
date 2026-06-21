import { NextRequest } from 'next/server';

import { handleAdminSubscription } from '../../../../lib/server/settings-readonly';

export function GET(request: NextRequest) {
  return handleAdminSubscription(request);
}
