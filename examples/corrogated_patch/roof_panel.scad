include <BOSL2/std.scad>
include <BOSL2/walls.scad>

width  = 90;   // pitch (peak-to-peak)
height = 25;   // peak-to-trough
total_len = 300;
patchLength = 100; 
// Flats at the top and bottom of each corrugation (must be <= width)
crest_flat  = 30;
trough_flat = 30;

// Smoothness: points per wave (higher = smoother slopes)
samples_per_wave = 20;

// ---- trapezoidal corrugation with flat crest/trough ----
amp = height/2;    // amplitude

// Piecewise trapezoid wave: flat top -> down slope -> flat bottom -> up slope
function wave_y(u, p, amp, ft, fb) =
    let(ws = max(0, (p - ft - fb)/2))   // width of each slope
    // Handle degenerate case where there's no slope:
    (ws == 0) ? ( (u <= ft) ? amp : (u <= ft + fb) ? -amp : amp )
    :
    (u <= ft) ? amp :
    (u <= ft + ws) ?           // down slope
        amp - ( (u - ft)/ws )*(2*amp) :
    (u <= ft + ws + fb) ?      // flat bottom
        -amp :
    (u <= ft + 2*ws + fb) ?    // up slope
        -amp + ( (u - (ft + ws + fb))/ws )*(2*amp) :
        amp;                   // back to flat top

// Build the path along +X using the trapezoid wave in Y
function corrugated_path(len, pitch, amp, ft, fb, spw) =
    let(nwaves = len/pitch,
        steps  = max(2, floor(spw*nwaves)))
    [ for (i = [0:steps])
        let(x = i*(len/steps),
            u = x % pitch,
            y = wave_y(u, pitch, amp, ft, fb))
        [x, y, 0]
    ];

path = corrugated_path(total_len, width, amp, crest_flat, trough_flat, samples_per_wave);

// Sweep your profile along the corrugated path.
// Centering the profile keeps it around the path’s centerline.
path_extrude(path)
    square([patchLength,3], center=true, $fn=6);
