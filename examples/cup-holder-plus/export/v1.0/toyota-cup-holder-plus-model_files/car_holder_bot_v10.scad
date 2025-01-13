$fn = 250;


// IMPORTAINT - ENSURE THESE MATCH YOU Cup holder below (toyata rav below)
$in_car_cup_holder_height = 65;
$in_car_holder_top_diameter = 74;
$in_car_bottom_diameter = 66.5;



$magnet_num=4;
$magnet_hole_diameter=2.2;
$magnet_hole_depth=2.2;
$magnet_hole_position=32.2;
$magnet_hole_offset=$in_car_cup_holder_height-$magnet_hole_depth;

$holder_length=225;
$holder_width=110;
$holder_depth=100;


// Cup holder holes Configuration
$cup_holder_center_offset= 43;
$cup_holder_y_offset = 0;
$cup_holder_height=120;
$cup_holder_top_radius=93;
$cup_holder_botton_radius=75;
$cup_holder_floor_depth=10;


// Phone Holder Configuration
$phone_holder1_offset=-105;
$phone_holder2_offset=80;

$phone_holder_width=23;
$phone_holder_floor_depth=12;
$phone_holder_depth=100;
$phone_holder_length=100;



$center_spacing_size = [190, 60, 90];

// Center Spacing Square Configration
$center_spacing_location = [-90, -$center_spacing_size[1]/2 , $in_car_cup_holder_height+$phone_holder_floor_depth+10 ];



// (unused) side cutout
$side_cutout_1_width = 95;
$side_cutout_height = 50;
$side_cutout_x_offset = 23;
$side_cutout_y_offset = $in_car_cup_holder_height + 20;




// Mode: simple or rounded ('simple' for making changes/during development - 'rounded' to render with curves)
$mode = "simple";

module roundedCube(size = [10, 10, 10], radius = 2, center = false) {
    // Calculate adjustment for Minkowski thickness
    extra = $mode == "simple" ? 0 : radius;
    translate(center ? -[size[0]/2 + extra, size[1]/2 + extra, size[2]/2 + extra] : [extra, extra, extra])
    
    if ($mode == "simple") {
        cube(size, center = false);
    } else if ($mode == "rounded") {
        minkowski() {
            #cube(size - [2 * radius, 2 * radius, 2 * radius], center = false);
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
        echo("Invalid mode. Use 'simple' or 'rounded'.");
    }
}


//translate([40, 0, 0])
//difference(){
//    cylinder(h=$in_car_cup_holder_height,d1=$bottom_diameter, d2=$top_diameter);
 //   translate([0,0,1]) cylinder(h=$in_car_cup_holder_height,d1=$bottom_diameter-3, d2=$bottom_diameter-2);
//	for (i=[0:360/$magnet_num:360])
//		rotate([0,0,i])translate([0,$magnet_hole_position,$magnet_hole_offset]) cylinder($magnet_hole_depth, d=$magnet_hole_diameter);
//}



translate([20, 0, 0])
cylinder(h=$in_car_cup_holder_height,d1=$in_car_bottom_diameter, d2=$in_car_holder_top_diameter);


difference(){
    // Main Holder Box
    translate([-$holder_length/2, -$holder_width/2, $in_car_cup_holder_height]) roundedCube(size = [$holder_length, $holder_width, $holder_depth], radius = 10);
    
    // Side Through cutout 1
  //  translate([$side_cutout_x_offset-15, -90, $side_cutout_y_offset])
  //  roundedCube([$side_cutout_1_width, 200, $side_cutout_height]);
    
    // Side Through cutout 2
  //  translate([-$side_cutout_x_offset-80, -100, $side_cutout_y_offset])
   // roundedCube([$side_cutout_1_width, 200, $side_cutout_height]);
    
    
    // End-To-End Through cutout
   // translate([-150, -35, 100])
   // roundedCube([2000, 90, 50]);
    
    
    
    // Cup holder 1
   translate([$cup_holder_center_offset, $cup_holder_y_offset, $in_car_cup_holder_height+$cup_holder_floor_depth]) roundedCylinder(h=$cup_holder_height, d1=$cup_holder_botton_radius, d2=$cup_holder_top_radius);
    
     // Cup holder 2
    translate([-$cup_holder_center_offset, $cup_holder_y_offset, $in_car_cup_holder_height+$cup_holder_floor_depth]) roundedCylinder(h=$cup_holder_height, d1=$cup_holder_botton_radius, d2=$cup_holder_top_radius);
    
    // Phone Holder 1 
    translate([$phone_holder1_offset, -$phone_holder_depth/2, $in_car_cup_holder_height+$phone_holder_floor_depth])
    roundedCube([$phone_holder_width, $phone_holder_length, $phone_holder_depth], radius=10);
      
     // Phone Holder 2
    translate([$phone_holder2_offset, -$phone_holder_depth/2, $in_car_cup_holder_height+$phone_holder_floor_depth])
    roundedCube([$phone_holder_width, $phone_holder_length, $phone_holder_depth], radius=10);
    
    // Center Phone Cutout
    translate($center_spacing_location)
    roundedCube($center_spacing_size);
    
    // Cutout from cup to phone 1
 //   translate([55, -25, $in_car_cup_holder_height+$phone_holder_floor_depth])
 //   roundedCube([30, 50, 100]);
    
    
    // Cutout from cup to phone 2
 //   translate([-100, -25, $in_car_cup_holder_height+$phone_holder_floor_depth])
 //   roundedCube([25, 50, 100]);
};

