/*

    "Tough Case INSTA360 X2" with slightly bigger holes


    Based on "Tough Case INSTA360 X2" from https://makerworld.com/en/models/200043#profileId-220576 by https://makerworld.com/en/@gianluigideruvo
*/
$fn = 200;

cylinderWidth = 3;

cylinderRotation = 90;
all_the_way = 1000;
cylinderXOffset = 27;
cylinderZOffset = 60;

sideCylinderXOffset = 31;


centerCaseSTLCoords = [-128, 200, -12.3];

difference(){
    
    translate(centerCaseSTLCoords)
    import("./source/Case_INSTA360_lens_cap_piece3.stl");



// Top cylinder    
     translate([-400, 2, 0])
        rotate([0, cylinderRotation, 0])
    cylinder(h=all_the_way, r=cylinderWidth);
    
    // side cylinder    
     translate([sideCylinderXOffset, all_the_way, 0])
        rotate([90, 0, 0])
    cylinder(h=all_the_way, r=cylinderWidth);
    
        // side cylinder    
     translate([-sideCylinderXOffset, all_the_way, 0])
        rotate([90, 0, 0])
    cylinder(h=all_the_way, r=cylinderWidth);
    
    // side re-enforcement

    
    
    
  /*  translate([cylinderXOffset, 0, cylinderZOffset])

    
    
    translate([-cylinderXOffset, 0, cylinderZOffset])
    rotate([0, -cylinderRotation, 0])
    #cylinder(h=all_the_way, r=cylinderWidth);*/
 }

        rotate([90,0,90])
        translate([20,-5,35])
       linear_extrude(1)
       square([20,10]);
       
        rotate([90,0,90])
        translate([20,-5,-35])
       linear_extrude(1)
       square([20,10]);