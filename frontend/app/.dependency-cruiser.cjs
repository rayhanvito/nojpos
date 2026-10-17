// dependency-cruiser — batas lapis FE (dokumen arsitektur frontend NOJPOS).
// Aturan menegakkan keadaan saat ini; pelanggaran lama yang belum
// diperbaiki dicatat di .dependency-cruiser-known-violations.json
// (baseline, dibuat dengan --baseline) dan diabaikan lewat --ignore-known
// pada script npm "depcruise". Pelanggaran BARU tetap gagal.
/** @type {import("dependency-cruiser").IConfiguration} */
module.exports = {
  forbidden: [
    {
      name: "primitives-no-app-layers",
      comment:
        "Lapis 1 (shadcn primitives, src/shared/ui/primitives) tidak boleh mengimpor lapis aplikasi shared/ui (*.tsx di src/shared/ui), pages, offline, atau components.",
      severity: "error",
      from: { path: "^src/shared/ui/primitives/" },
      to: { path: "^src/(shared/ui/[^/]+\.tsx|pages/|offline/|components/)" },
    },
    {
      name: "shared-no-pages",
      comment: "src/shared tidak boleh mengimpor src/pages.",
      severity: "error",
      from: { path: "^src/shared/" },
      to: { path: "^src/pages/" },
    },
    {
      name: "offline-only-from-offline-main",
      comment:
        "src/offline hanya boleh diimpor oleh src/offline-main.tsx (dan sesama file di src/offline); aplikasi utama tidak boleh mengimpor offline.",
      severity: "error",
      from: { pathNot: "^(src/offline-main\.tsx|src/offline/)" },
      to: { path: "^src/offline/" },
    },
    {
      name: "features-no-pages-offline",
      comment:
        "src/features boleh mengimpor src/shared, bukan src/pages atau src/offline.",
      severity: "error",
      from: { path: "^src/features/" },
      to: { path: "^src/(pages/|offline/)" },
    },
    {
      name: "shared-no-features",
      comment: "src/shared tidak boleh mengimpor src/features (arah lapis: features -> shared).",
      severity: "error",
      from: { path: "^src/shared/" },
      to: { path: "^src/features/" },
    },
    {
      name: "pages-no-offline",
      comment:
        "pages boleh mengimpor shared & components, bukan offline (aturan (d) dokumen; tumpang tindih dengan offline-only-from-offline-main agar eksplisit).",
      severity: "error",
      from: { path: "^src/pages/" },
      to: { path: "^src/offline/" },
    },
  ],
  options: {
    doNotFollow: { path: "node_modules" },
    tsConfig: { fileName: "tsconfig.json" },
    enhancedResolveOptions: {
      exportsFields: ["exports"],
      conditionNames: ["import", "require", "node", "default", "types"],
      extensions: [".ts", ".tsx", ".js", ".jsx", ".css", ".json"],
    },
  },
};
