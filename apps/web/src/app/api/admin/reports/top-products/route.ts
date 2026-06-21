import { NextRequest } from 'next/server';

import { handleAdminReport } from '../../../../../lib/server/reports';

export function GET(request: NextRequest) {
  return handleAdminReport(request, 'top-products');
}
