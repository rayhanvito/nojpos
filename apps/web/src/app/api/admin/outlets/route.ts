import { NextRequest } from 'next/server';

import { handleAdminOutlets } from '../../../../lib/server/outlets';

export function GET(request: NextRequest) {
  return handleAdminOutlets(request);
}
