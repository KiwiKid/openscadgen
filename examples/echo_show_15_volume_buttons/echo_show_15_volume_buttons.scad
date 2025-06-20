

	include <BOSL2/std.scad>;

	$fa = .01;
	$fs = $preview ? 5 : 1;
	$fn = 200;

    renderType = "";
        holderHeight = 28;
 minusOffset = -64; 
  plusOffset = minusOffset+22;
    holderLength = 160;
 holderWeight = 10;
 
module button_holder(diameter, holderWeight, holderHeight = 25){

up(holderHeight/2)
down(holderWeight/2)
union(){
    cyl(d=diameter, h=holderHeight, center=true, chamfer1=-4, chamfer2=-5, rounding=2);
  
    }
}

	module echo_show_15_volume_buttons(){
    
 
  
    back(holderHeight/3)
		difference(){
			cuboid([holderLength,holderHeight,holderWeight], rounding=2,  edges=[TOP]);
            down(4)
			cuboid([holderLength+2,holderHeight*0.8,holderWeight]);
         //   down(7)
          // cuboid([holderLength+2,holderHeight*1.5,holderWeight]);
            // swing_text_up("ALEXA - +");
            
         up(holderWeight/2*1.2-0.01)
            right(minusOffset)
            button_holder(diameter=11, holderWeight=holderWeight, holderHeight=30);
            
          up(holderWeight/2*1.2-0.01)
            right(plusOffset)
            button_holder(diameter=11, holderWeight=holderWeight, holderHeight=30);
		}
        
        difference(){
        
        union(){
            back(holderHeight/3)
           // back(holderWeight/2)
            right(minusOffset)
            button_holder(diameter=10, holderWeight=holderWeight);
            
            back(holderHeight/3)
          //  back(holderWeight/2)
            right(plusOffset)
            button_holder(diameter=10, holderWeight=holderWeight);
        
        }
            // Top button labels
            textWidthOffset = -3;
            back(3)
               right(plusOffset)
               translate([0, textWidthOffset, holderHeight-8.5])  
               linear_extrude(height=3)
               text("+", size=26, halign="center", valign="bottom", font="Courier");
               
               right(minusOffset)
               translate([0, 10+textWidthOffset, holderHeight-8.5])
               linear_extrude(height=3)
               text("-", size=26, halign="center", valign="bottom", font="Courier");
           }
       }    
        
        module swing_text_up(text_str) {
    // Parameters
    text_size = 20;
    text_height = 1;

            for (a = [-14 : 0.5 : 90]) {
                // Shift text so bottom edge is at origin (hinge line)
                translate([0, 0, 0])
                    rotate([a, 0, 0])
                        translate([0, 0, 0])  // pivot point at bottom edge
                            linear_extrude(height=text_height)
                                text(text_str, size=text_size, halign="center", valign="bottom");
            }
        }

        
        difference(){
        right(2.5)
            fwd(4)
                up(6)
                union(){
                right(30)
            swing_text_up("ALEXA");
            left(55)
             swing_text_up("– +");
             
                                  
             
           
            }
        
        
        back(holderHeight/3)
            right(minusOffset)
            button_holder(diameter=11, holderWeight=holderWeight, holderHeight=holderHeight);
            
            back(holderHeight/3)
            right(plusOffset)
            button_holder(diameter=11, holderWeight=holderWeight, holderHeight=holderHeight);
        }
        
        


	


    sliced(renderType=renderType) {
        echo_show_15_volume_buttons();
    }
       








	
     
module sliced(
    renderType = "horzSlice",        // "horzSlice", "vertSlice", or "all"
    sliceSize = 1000,
    sliceThickness = 0.3,
    showRawSlices = false,
    horzSlicePos = [-500, -500, 7],
    vertSlicePos = [-3, -500, -500]
) {
   
    module horz_slice(raw=false) {
        if (raw) {
        rotate([0,90,0])
            translate(horzSlicePos)
                cube([sliceSize, sliceSize, sliceThickness], center=false);
        } else {
            intersection() {
                children();
                rotate([0,90,0])
                translate(horzSlicePos)
                    cube([sliceSize, sliceSize, sliceThickness], center=false);
            }
        }
    }

    module vert_slice(raw=false) {
        if (raw) {
        rotate([0,0,90])
            translate(vertSlicePos)
                cube([sliceThickness, sliceSize, sliceSize], center=false);
        } else {
         
            intersection() {
                children();
                  rotate([0,0,90])
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
        #vert_slice(raw=true);
        // show full object
        children();
    } else {
        // show full object
        children();
    }
}

