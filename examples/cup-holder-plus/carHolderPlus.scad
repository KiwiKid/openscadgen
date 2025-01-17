$fn = 20;
/*
    A cup holder upgrade for the toyota rav4 (v1.4)

    Includes: 
    - Two options for the holder size (twoLargerHolders, oneLargerHolderOneSmallerHolder)
    - Phone holders at either end
    - Adjustable cup holder size


Note: This design is a learning tool for openSCAD, while the design is functional, the code should not be used as a reference for best (or even good) practices for writing openSCAD code (includes hardcoded values, hidden dependencies between variables, non-modular etc.)

*/




// oneLargerHolderOneSmallerHolder, twoLargerHolders
cup_holders_mode = "twoLargerHolders";


cup_holder_y_offset = 2.5;
cup_holder_center_offset= 43;
cup_holder_height=120;
cup_holder_floor_depth=10;

cup_holder_1_top_radius=95;
cup_holder_1_botton_radius=95;

cup_holder_2_top_radius=85;
cup_holder_2_botton_radius=75;

phone_holder1_offset=-110;
phone_holder2_offset=90;

phone_holder_width=23;


phone_holder_floor_depth=12;
phone_holder_height=100;
phone_holder_depth=100;


// Mode: simple or rounded - (this is redundent in newer builds of openSCAD, as at Jan 2025, the development build (with peformance improvments, i.e. manfold) will run this model, with minkowski transform without lagging
$mode = "rounded";

// IMPORTANT - Ensure these match your cup holder (toyata rav4 set below)
in_car_cup_holder_height = 56;
in_car_cup_holder_top_diameter = 78;
in_car_cup_holder_bottom_diameter = 66.5;
in_car_cup_holder_center_offset = 16;


holder_length=240;
holder_width=110;
holder_depth=100;



module roundedCube(size = [10, 10, 10], radius = 2, center = false) {
    // Calculate adjustment for Minkowski thickness
    extra = $mode == "simple" ? 0 : radius;
    translate(center ? -[size[0]/2 + extra, size[1]/2 + extra, size[2]/2 + extra] : [extra, extra, extra])
    
    if ($mode == "simple") {
        cube(size, center = false);
    } else if ($mode == "rounded") {
        minkowski() {
            cube(size - [2 * radius, 2 * radius, 2 * radius], center = false);
            sphere(radius);
        }
    } else {
        echo("Invalid mode. Use 'simple' or 'rounded'.");
    }
}

module roundedCylinder(h = 20, d1 = 10, d2 = 10, radius = 2, center = false) {
    // Compute outer dimensions (same for both modes)
    outer_h = h + 2 * radius;
    outer_d1 = d1 + 2 * radius;
    outer_d2 = d2 + 2 * radius;

    // Translate based on outer dimensions for consistent placement
    translate(center ? -[max(outer_d1, outer_d2) / 2, max(outer_d1, outer_d2) / 2, outer_h / 2] : [0, 0, 0])
    
    if ($mode == "simple") {
        cylinder(h = h, d1 = d1, d2 = d2, center = false);
    } else if ($mode == "rounded") {
        minkowski() {
            cylinder(h = h, d1 = d1, d2 = d2, center = false);
            sphere(radius);
        }
    } else {
     //   echo("Invalid mode. Use 'simple' or 'rounded'.");
    }
}


module cupHolder(){

    translate([in_car_cup_holder_center_offset, 0, 0])
    cylinder(h=in_car_cup_holder_height,d1=in_car_cup_holder_bottom_diameter, d2=in_car_cup_holder_top_diameter);
    
        


    difference(){
        // Main Holder Box
        translate([-holder_length/2, -cup_holder_1_top_radius/2-5, in_car_cup_holder_height]) roundedCube(size = [holder_length, holder_width, holder_depth], radius = 10);
        
        // Cup holder 1
       translate([cup_holder_center_offset, cup_holder_y_offset, in_car_cup_holder_height+cup_holder_floor_depth]) roundedCylinder(h=cup_holder_height, d1=cup_holder_1_botton_radius, d2=cup_holder_1_top_radius);
        
        if (cup_holders_mode == "oneLargerHolderOneSmallerHolder") {
         // Cup holder 2 (small)
        translate([-cup_holder_center_offset, cup_holder_y_offset, in_car_cup_holder_height+cup_holder_floor_depth]) roundedCylinder(h=cup_holder_height, d1=cup_holder_2_botton_radius, d2=cup_holder_2_top_radius);
        } else if (cup_holders_mode == "twoLargerHolders") {
           // Cup holder 2 (big)

        translate([-cup_holder_center_offset, cup_holder_y_offset, in_car_cup_holder_height+cup_holder_floor_depth]) roundedCylinder(h=cup_holder_height, d1=cup_holder_1_botton_radius, d2=cup_holder_1_top_radius);
        
        } else {
           echo("<b style='color:red'>", A="Failed");
        
        }
        
        // Phone Holder 1 
        translate([phone_holder1_offset, -47, in_car_cup_holder_height+phone_holder_floor_depth])
        roundedCube([phone_holder_width, phone_holder_depth, phone_holder_height], radius=10);
        
        

          
         // Phone Holder 2
        translate([phone_holder2_offset, -48, in_car_cup_holder_height+phone_holder_floor_depth])
        roundedCube([phone_holder_width, phone_holder_depth, phone_holder_height], radius=10);
        

        
        
        
        // Soften some edges in the middle
        if (cup_holders_mode == "twoLargerHolders") {
        // Center Phone Cutout
        translate([5, 5, in_car_cup_holder_height+60])
        roundedCube([180, 60, 90], center=true);
        } else if (cup_holders_mode == "oneLargerHolderOneSmallerHolder") {
         
         translate([-90, -20, in_car_cup_holder_height+phone_holder_floor_depth+10])
        
        roundedCube([200, 50, 100]);
        
        
        translate([phone_holder2_offset, 8, phone_holder_depth+phone_holder_floor_depth+20])
        roundedCube([20, 60, 90], center=true);
        }
        
    };

};




cupHolder();

        