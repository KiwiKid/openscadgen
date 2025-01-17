include <../../library/BOSL2/std.scad>;


/*

    "Tough Case INSTA360 X2" with slightly bigger holes


    Based on "Tough Case INSTA360 X2" from https://makerworld.com/en/models/200043#profileId-220576 by https://makerworld.com/en/@gianluigideruvo
*/
$fn = 20;

cylinderWidth = 3.5;

cylinderRotation = 7;
all_the_way = 1000;
cylinderXOffset = 26.6;
cylinderZOffset = 60;


centerCaseSTLCoords = [-435, -127.67, 0];

includeGoProMount = true;

$fa = 2;
$fs = 0.25;
Extra_Mount_Depth = 4;

module nut_hole()
{
	rotate([0, 90, 0]) // (Un)comment to rotate nut hole
	rotate([90, 0, 0])
		for(i = [0:(360 / 3):359])
		{
			rotate([0, 0, i])
				cube([4.6765, 8.1, 5], center = true);
		}
}

module flap(Width)
{
	rotate([90, 0, 0])
	union()
	{
		translate([3.5, (-7.5), 0])
			cube([4 + Extra_Mount_Depth, 15, Width]);


		translate([0, (-7.5), 0])
			cube([7.5 + Extra_Mount_Depth, 4, Width]);

		translate([0, 3.5, 0])
			cube([7.5 + Extra_Mount_Depth, 4, Width]);

		difference()
		{
			cylinder(h = Width, d = 15);

			translate([0, 0, (-1)])
				cylinder(h = Width + 2, d = 6);
		}
	}
}


module mount2()
{
	union()
	{

			translate([0, 4, 0])
		flap(3);

		translate([0, 10.5, 0])
		flap(3);
	}
}

module mount3()
{
	union()
	{
		difference()
		{
			translate([0, (-2.5), 0])
				flap(8);

			translate([0, (-8.5), 0])
				nut_hole();
		}

		mount2();
	}
}

mountRotation = [0,-10,90];
mountLocation = [-20,-5.8,40];

mountJoinerSize = [28,30];
mountJoinerRotation = [90,0,0];
mountJoinerLocation = [-mountJoinerSize[0]/2+14, 34.6, 10.3];

//mountStabilizerSize = [15, 14];
//mountStabilizerExtention = 10;
//mountStabilizerRoation = [101,0,0];
//mountStabilizerLocation = [-mountJoinerSize[0]/2+14, 50.5, 4];

difference(){
    
    translate(centerCaseSTLCoords)
    import("./source/Case_INSTA360_lens_cap_piece2.stl");

    translate([cylinderXOffset, 0, cylinderZOffset])
    rotate([0, cylinderRotation, 0])
    cylinder(h=all_the_way, r=cylinderWidth);
    
    
    translate([-cylinderXOffset, 0, cylinderZOffset])
    rotate([0, -cylinderRotation, 0])
    cylinder(h=all_the_way, r=cylinderWidth);
    
    
   
 }
 
 
 if (includeGoProMount) {
 
 
 translate([0, 0, -10]){
 
        rotate(mountRotation)
        translate(mountLocation)
     mount2();

     rotate(mountJoinerRotation)
     translate(mountJoinerLocation)
     linear_extrude(height = 7)
     squircle(mountJoinerSize);

    // rotate(mountStabilizerRoation)
    // translate(mountStabilizerLocation)
    // linear_extrude(height= 10)
   //  hull(){
    //    #squircle(mountStabilizerSize);
       // translate([0,-20,40]) {
        //    #squircle(mountStabilizerSize);
        //};
    //   }

    }
     
  //   text(text="GC FPV", size =10)
        
}