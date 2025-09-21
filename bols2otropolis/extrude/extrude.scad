// Extrude Examples - Linear and Rotate Extrude with BOSL2
// Demonstrates various parameters and use cases

include <BOSL2/std.scad>

// Parameters
extrude_type = "linear"; // "linear" or "rotate"
shape_type = "circle"; // "circle", "square", "star", "polygon"
height = 20;
scale_factor = 1.0;
twist = 0;
slices = 0;
convexity = 10;

// Shape generation functions
function circle_points(r, $fn=16) = [for(i=[0:$fn-1]) [r*cos(i*360/$fn), r*sin(i*360/$fn)]];

function square_points(size) = [
    [-size/2, -size/2],
    [size/2, -size/2], 
    [size/2, size/2],
    [-size/2, size/2]
];

function star_points(outer_r, inner_r, points=5) = [
    for(i=[0:points*2-1]) 
        let(angle = i * 180/points)
        let(r = i%2==0 ? outer_r : inner_r)
        [r*cos(angle), r*sin(angle)]
];

function polygon_points(sides, radius) = [
    for(i=[0:sides-1]) 
        let(angle = i * 360/sides)
        [radius*cos(angle), radius*sin(angle)]
];

// Get shape points based on type
function get_shape_points(type, size) = 
    type == "circle" ? circle_points(size/2) :
    type == "square" ? square_points(size) :
    type == "star" ? star_points(size/2, size/4) :
    type == "polygon" ? polygon_points(6, size/2) :
    circle_points(size/2);

// Main extrude function
module extrude_shape() {
    points = get_shape_points(shape_type, 20);
    
    if (extrude_type == "linear") {
        linear_extrude(
            height = height,
            scale = scale_factor,
            twist = twist,
            slices = slices,
            convexity = convexity
        ) {
            polygon(points);
        }
    } else if (extrude_type == "rotate") {
        rotate_extrude(convexity = convexity) {
            translate([10, 0, 0]) {
                polygon(points);
            }
        }
    }
}

// Example 1: Basic linear extrude
module basic_linear_extrude() {
    linear_extrude(height = 20) {
        circle(d = 20);
    }
}

// Example 2: Linear extrude with scaling
module scaled_linear_extrude() {
    linear_extrude(height = 30, scale = 0.5) {
        circle(d = 20);
    }
}

// Example 3: Linear extrude with twist
module twisted_linear_extrude() {
    linear_extrude(height = 40, twist = 180) {
        square([15, 3], center = true);
    }
}

// Example 4: Linear extrude with both scale and twist
module complex_linear_extrude() {
    linear_extrude(height = 50, scale = 0.2, twist = 360) {
        polygon(star_points(10, 4, 5));
    }
}

// Example 5: Basic rotate extrude
module basic_rotate_extrude() {
    rotate_extrude() {
        translate([15, 0, 0]) {
            circle(d = 10);
        }
    }
}

// Example 6: Rotate extrude with different shapes
module shaped_rotate_extrude() {
    rotate_extrude() {
        translate([20, 0, 0]) {
            square([8, 4], center = true);
        }
    }
}

// Example 7: Rotate extrude creating a torus
module torus_rotate_extrude() {
    rotate_extrude() {
        translate([25, 0, 0]) {
            circle(d = 10);
        }
    }
}

// Example 8: Rotate extrude with star profile
module star_rotate_extrude() {
    rotate_extrude() {
        translate([18, 0, 0]) {
            polygon(star_points(4, 1.5, 5));
        }
    }
}

// Example 9: Linear extrude with polygon
module polygon_linear_extrude() {
    linear_extrude(height = 25, scale = 1.5) {
        polygon([
            [0, 0], [10, 0], [15, 10], [5, 15], [-5, 10]
        ]);
    }
}

// Example 10: Rotate extrude with complex profile
module complex_rotate_extrude() {
    rotate_extrude() {
        translate([22, 0, 0]) {
            polygon([
                [0, 0], [6, 0], [8, 2], [6, 4], [0, 4]
            ]);
        }
    }
}

// Main module - choose which example to show
module main() {
    if (extrude_type == "linear") {
        if (shape_type == "circle") basic_linear_extrude();
        else if (shape_type == "square") scaled_linear_extrude();
        else if (shape_type == "star") complex_linear_extrude();
        else if (shape_type == "polygon") polygon_linear_extrude();
        else twisted_linear_extrude();
    } else {
        if (shape_type == "circle") basic_rotate_extrude();
        else if (shape_type == "square") shaped_rotate_extrude();
        else if (shape_type == "star") star_rotate_extrude();
        else if (shape_type == "polygon") complex_rotate_extrude();
        else torus_rotate_extrude();
    }
}

main();
