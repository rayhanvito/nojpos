<?php

return [
    'checkout' => [
        // Stage-2 production default: checkout is quote-only. Set NOJPOS_CHECKOUT_REQUIRE_QUOTE=false
        // only for controlled compatibility tests or a temporary legacy rollout window.
        'require_quote_for_checkout' => env('NOJPOS_CHECKOUT_REQUIRE_QUOTE', true),
    ],
];
