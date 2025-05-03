

	include <BOSL2/std.scad>;

	$fa = .01;
	$fs = $preview ? 5 : 1;
	$fn = 200;
    
    renderType = "horzSlice";


	module washing_machine_seal_v3(){
        thickness =1.5;
		outerRadius = 160;
        innerRadius = outerRadius-thickness;
        height = 35;
        
        gapHeight = height*0.38;
        
        cutTranslate = [-10,0,gapHeight];
        
        cut2Translate = [230,-10,gapHeight];
        
        cutWidth = 2;
        
        difference(){
         intersection() {
        union(){
        tube(ir=innerRadius, or=outerRadius, h=height, chamfer=2, rounding2=2, orounding2=2);

            

            }
            
            // Wedge mask: 1/3 circle (120 degrees)
            rotate([0,0,-60])  // Center the wedge
                linear_extrude(height)
                    arc(r=outerRadius + 2, angle=150); // slightly oversized radius to ensure full coverage
        }
        
        translate(cutTranslate)
        up(cutWidth/2)
        cuboid([200,400,cutWidth]);
        
        rotate([0, 0,60])
        translate(cut2Translate)
        up(50)
        cuboid([200,200,100]);
        
     /* rotate([0, 0,60])
        up(cutWidth/2)
        translate(cut2Translate)
        cuboid([230,230,1]);
          
        translate([-180,100,gapHeight+1])
        rotate([180, 0,60])
//        translate([230,-10,gapHeight])
        #wedge([300,299,2]);
*/
        }
        
        
        // Rim Circle water-keeper-in
        difference(){
         intersection() {
            up(height/2-1)
                tube(ir=innerRadius-1, or=outerRadius, h=2.5, rounding=1.25);
            
            // Wedge mask: 1/3 circle (120 degrees)
            rotate([0,0,-60])  // Center the wedge
                linear_extrude(height)
                    arc(r=outerRadius + 2, angle=150); // slightly oversized radius to ensure full coverage
        }
        
        cube_width = 200;
        // non-active rim-preventer
        translate([70,-140-(cube_width/2),0])
        #cube([200,cube_width,100]);
        }
        
    }


    sliced(renderType=renderType) {
        washing_machine_seal_v3();
    }
       

       
       
       
       
       
       
       
       
       
       
       
     
module sliced(
    renderType = "hh",        // "horzSlice", "vertSlice", or "all"
    sliceSize = 1000,
    sliceThickness = 8,
    showRawSlices = false,
    horzSlicePos = [-500, -500, 9.3],
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

