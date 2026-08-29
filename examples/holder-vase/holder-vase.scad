include <BOSL2/std.scad>;

$fa = .01;
$fs = $preview ? 5 : 1;
$fn = 200;

include <BOSL2/skin.scad>;
include <BOSL2/paths.scad>;

height = 120;
wall_bump_rate = 200;
base_rotate = [0, 0, 0];

scaleAll = 0.3;

function make_circle_points(r, steps=40) = [
    for(i = [0:steps-1])
        let(angle = 360 * i / steps)
        [r * cos(angle), r * sin(angle)]
];

function get_profile(step, total_steps, base_r=26, taper=3, height=40) =
    let(
        t = step / total_steps,
        r = base_r - t * taper + sin(step * 5) * 4.5,
        x = sin(step * 10) * 10,
        y = cos(step * 2) * 3,
        z = t * height
    )
    [x, y, z];

module vase(wall_thickness=2, floor_thickness=50, wall_bump_rate=40, height=60) {
    // Only allow heights >= 100
   // actual_height = max(height, 100);
    
    core_radius = height;
    path = [for(i = [0:40]) get_profile(i, 14, height=height)];
    radii = [
        for(i = [0:40])
            let(
                t = i / 40,
                r = core_radius - t * 0.5 + sin(i * wall_bump_rate/2) * 10.5
            )
            r
    ];
    
    // Get the first point and radius for the base
    first_point = path[0];
    base_radius = radii[0];
    
    // Get the last point and radius for the top
    last_point = path[len(path)-1];
    top_radius = radii[len(radii)-1];
    
    // Create the hollow vase with integrated floor
    
    wall_thick = 1.1;
    
    difference() {
        union() {
            
            // Main body
            path_sweep(
                make_circle_points(wall_thick), 
                path, 
                scale=radii,
                caps=true
            );
            
            rotate(base_rotate)
            // Floor as part of the main body
            translate([first_point.x, first_point.y, -30])
            cylinder(h=floor_thickness, r=base_radius*wall_thick);
        }
        
        
        // Inner shell (slightly smaller)
        translate([0, 0, floor_thickness])
        path_sweep(
            make_circle_points(1), 
            [for(p = path) [p.x, p.y, p.z - floor_thickness]], 
            scale=[for(r = radii) r - wall_thickness],
            caps=true
        );
        
        // Clean cut at the top using a cylinder
        translate([0, 0, last_point.z])
        rotate(base_rotate)
        cylinder(h=200, r=top_radius * 2, anchor=CENTER);
    }
}

// Only generate the taller vases
sliced(renderType="") {
scale(scaleAll)
    vase(wall_bump_rate=wall_bump_rate, height=height);
}

module sliced(
    renderType = "horzSlice",        // "horzSlice", "vertSlice", or "all"
    sliceSize = 1000,
    sliceThickness = 0.3,
    showRawSlices = false,
    horzSlicePos = [-500, -500, 0],
    vertSlicePos = [0, -500, -500]
) {
   
    module horz_slice(raw=false) {
        if (raw) {
            translate(horzSlicePos)
                cuboid([sliceSize, sliceSize, sliceThickness], anchor=[-1,-1,-1]);
        } else {
            intersection() {
                children();
                translate(horzSlicePos)
                    cuboid([sliceSize, sliceSize, sliceThickness], anchor=[-1,-1,-1]);
            }
        }
    }

    module vert_slice(raw=false) {
        if (raw) {
            translate(vertSlicePos)
                cuboid([sliceThickness, sliceSize, sliceSize], anchor=[-1,-1,-1]);
        } else {
            intersection() {
                children();
                translate(vertSlicePos)
                    cuboid([sliceThickness, sliceSize, sliceSize], anchor=[-1,-1,-1]);
            }
        }
    }

    if (renderType == "horzSlice") {
        horz_slice(raw=showRawSlices){
            children();
        }
    } else if (renderType == "vertSlice") {
        vert_slice(raw=showRawSlices){
            children();
        }
    } else if (renderType == "all") {
        // show raw slices for reference
        horz_slice(raw=true);
        vert_slice(raw=true);
        // show full object
        children();
    } else {
        // show full object
        children();
    }
}

