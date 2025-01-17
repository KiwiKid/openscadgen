/*

    "Tough Case INSTA360 X2" - Cap - with slightly bigger holes


    Based on "Tough Case INSTA360 X2" from https://makerworld.com/en/models/200043#profileId-220576 by https://makerworld.com/en/@gianluigideruvo
*/
$fn = 200;

cylinderWidth = 3.5;

includeCheckingCylinders = false;

//scaleFactor = cylinderWidth-1.5;

cylinderRotation = 90;
all_the_way = 1000;

        scaleFactor = cylinderWidth / 2.8;

    centerTopSTLCoords = [-128.5 * scaleFactor, -128 * scaleFactor, -12.3 * scaleFactor];

    translate(centerTopSTLCoords) {
        scale(cylinderWidth > 0 ?   [scaleFactor,scaleFactor, scaleFactor] : [1,1,1])
        import("./source/Case_INSTA360_lens_cap_piece1.stl");
    }
    
    if (includeCheckingCylinders){
        translate([0, 5, -100])
    rotate([0, 0, 0])
    #cylinder(h=all_the_way, r=cylinderWidth);
    
            translate([0, -3, -100])
    rotate([0, 0, 0])
    #cylinder(h=all_the_way, r=cylinderWidth);
  }