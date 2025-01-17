$fn = 20;
/*
    A cup holder upgrade sizer for the toyota rav4 (v1.4)

    Includes: 
    - Two options for the holder size (twoBig, oneBigOneSmall)
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




module cupHolderSizer(){

wall_thickness=1;
sizer_bottom_offset=10;

difference(){
    translate([in_car_cup_holder_center_offset, 0, 0])
    cylinder(h=in_car_cup_holder_height,d1=in_car_cup_holder_bottom_diameter, d2=in_car_cup_holder_top_diameter);
    
    // cut-out to save material 

    translate([in_car_cup_holder_center_offset, 0, wall_thickness])
    cylinder(h=in_car_cup_holder_height,d1=in_car_cup_holder_bottom_diameter-wall_thickness, d2=in_car_cup_holder_top_diameter-wall_thickness);
    

    // finger hole to get out

    translate([15, 100, in_car_cup_holder_height-30])
    rotate([90, 0, 0])
    #cylinder(h=1000,d=20);


}
}


cupHolderSizer();

        