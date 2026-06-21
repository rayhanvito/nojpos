import { NextRequest } from 'next/server';
import { handleAdminSession } from '../../../../lib/server/admin-session';

export async function GET(request: NextRequest) {
  return handleAdminSession(request);
}
