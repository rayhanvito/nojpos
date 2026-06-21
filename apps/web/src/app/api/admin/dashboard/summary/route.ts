import { NextRequest } from 'next/server';
import { handleAdminDashboardSummary } from '../../../../../lib/server/dashboard-summary';

export async function GET(request: NextRequest) {
  return handleAdminDashboardSummary(request);
}
