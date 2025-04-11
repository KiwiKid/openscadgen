
include <BOSL2/std.scad>;

	$fa = .01;
	$fs = $preview ? 5 : 1;
	$fn = 200;

	renderType = "all-all"; // horzSlice, vertSlice, all
    
    screwHolesUp = 10;
    screwHoles = [0, 5, 10, 15, 20, 25, 30];

	module hand_press_plate(plateWidth=100, plateHeight=100, plateAttachDepth=34, attachCyliderRadius=20, attachCyliderHeight=50, screwHoleDiameter=4, holderScrewDimpleDepth=3){
		
        difference(){
        union(){
            cyl(l=attachCyliderHeight, d=attachCyliderRadius, rounding=2);
            up(attachCyliderHeight-4)
            rotate([180,0,0])
            prismoid(size1=[plateWidth,plateHeight], size2=[attachCyliderRadius*0.8,attachCyliderRadius*0.8], h=plateAttachDepth, rounding=7);
        }
        
        for (pos = screwHoles) {
            rotate([90, 90, 0])
            translate([pos-screwHolesUp, 0, attachCyliderRadius/2 - 1])  // adjust as needed
            
            cyl(l=holderScrewDimpleDepth, d=screwHoleDiameter, rounding=1);
        }
        }
        
	}


	if (renderType == "horzSlice" || renderType == "all-all") {
		intersection(){
			hand_press_plate(); 
			fwd(500)
			left(500)
			
    #cube([1000,1000,0.3]);
		}
	}
    if(renderType == "vertSlice" || renderType == "all-all") {
		#intersection(){
			hand_press_plate();
			rotate([90,0,90])
			fwd(500)
			left(500)
			down(0)
			#cube([1000,1000,0.3]);
		}
	} 
   
   if(renderType == "all" || renderType == "all-all") {
		hand_press_plate();
	}
    
