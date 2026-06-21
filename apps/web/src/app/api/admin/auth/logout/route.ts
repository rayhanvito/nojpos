import { NextRequest } from 'next/server';
import { handleAdminLogout } from '../../../../../lib/server/admin-session';

export async function POST(request: NextRequest) {
  return handleAdminLogout(request);
}
