include <BOSL2/std.scad>;

$fa = .01;
$fs = $preview ? 5 : 1;
$fn = 200;

include <BOSL2/skin.scad>;
include <BOSL2/paths.scad>;

height = 120;
wall_bump_rate = 200;
base_rotate = [0, 10, 0];

scaleAll = 4;

function make_circle_points(r, steps=40) = [
    for(i = [0:steps-1])
        let(angle = 360 * i / steps)
        [r * cos(angle), r * sin(angle)]
];

function get_profile(step, total_steps, base_r=25, taper=3, height=40) =
    let(
        t = step / total_steps,
        r = base_r - t * taper + sin(step * 15) * 1.5,
        x = sin(step * 10) * 2,
        y = cos(step * 12) * 2,
        z = t * height
    )
    [x, y, z];

module vase(wall_thickness=2, floor_thickness=3, wall_bump_rate=50, height=40) {
    // Only allow heights >= 100
   // actual_height = max(height, 100);
    
    // Simplified path generation with less complex waves
    path = [for(i = [0:40]) get_profile(i, 40, height=height)];
    radii = [
        for(i = [0:40])
            let(
                t = i / 40,
                r = 30 - t * 8 + sin(i * wall_bump_rate/2) * 1.5
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
    
    difference() {
        union() {
            
            // Main body
            path_sweep(
                make_circle_points(1), 
                path, 
                scale=radii,
                caps=true
            );
            
            rotate(base_rotate)
            // Floor as part of the main body
            translate([first_point.x, first_point.y, -3])
            cylinder(h=floor_thickness, r=base_radius);
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
        cylinder(h=20, r=top_radius * 2, center=true);
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
                cube([sliceSize, sliceSize, sliceThickness], center=false);
        } else {
            intersection() {
                children();
                translate(horzSlicePos)
                    cube([sliceSize, sliceSize, sliceThickness], center=false);
            }
        }
    }

    module vert_slice(raw=false) {
        if (raw) {
            translate(vertSlicePos)
                cube([sliceThickness, sliceSize, sliceSize], center=false);
        } else {
            intersection() {
                children();
                translate(vertSlicePos)
                    cube([sliceThickness, sliceSize, sliceSize], center=false);
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

